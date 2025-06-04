package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	logger "github.com/sirupsen/logrus"
)

// Struct to capture the seal status response from Vault
type SealStatus struct {
	Sealed bool `json:"sealed"`
}

// Struct for unseal request payload
type UnsealRequest struct {
	Key string `json:"key"`
}

// Function to check and unseal a single Vault server
func checkAndUnsealVault(server string, unsealKeys []string, sealed *bool) {

	// load tls certificates
	clientTLSCert, err := tls.LoadX509KeyPair("tls/server.crt", "tls/server.key")
	if err != nil {
		logger.Error("Error loading certificate and key file:", err)
		return
	}
	caCert, err := os.ReadFile("tls/ca.crt")
	if err != nil {
		logger.Error("Error reading CA file:", err)
		return
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
				Certificates:       []tls.Certificate{clientTLSCert},
				InsecureSkipVerify: true, // this is needed, otherwise we fail due to:
				//   "msg": "Error fetching seal status:Get \"https://192.168.124.173:8200/v1/sys/seal-status\": tls: failed to verify certificate: x509: certificate is valid for 127.0.0.1, not 192.168.124.173",
			},
		},
	}

	resp, err := client.Get(server + "/v1/sys/seal-status")
	if err != nil {
		logger.Error("Error fetching seal status:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading response body:", err)
		return
	}
	var status SealStatus
	err = json.Unmarshal(body, &status)
	if err != nil {
		logger.Error("Error unmarshaling JSON:", err)
		return
	}

	if status.Sealed {
		*sealed = true
		// report the following each time, as showing attempts is more comforting than only showing once for a state change
		logger.Info(server, " is sealed. Attempting to unseal...")
		for _, key := range unsealKeys {
			jsonData := UnsealRequest{Key: key}
			jsonValue, _ := json.Marshal(jsonData)

			unsealResp, err := client.Post(server+"/v1/sys/unseal", "application/json", bytes.NewBuffer(jsonValue))
			if err != nil {
				logger.Error("Error posting unseal request:", err)
				return
			}
			unsealResp.Body.Close()

			// Check if unseal was successful by re-checking the seal status
			checkResp, err := client.Get(server + "/v1/sys/seal-status")
			if err != nil {
				logger.Error("Error re-checking seal status:", err)
				return
			}
			defer checkResp.Body.Close()

			body, err := io.ReadAll(checkResp.Body)
			if err != nil {
				logger.Error("Error reading check response body:", err)
				return
			}
			err = json.Unmarshal(body, &status)
			if err != nil {
				logger.Error("Error unmarshaling check response JSON:", err)
				return
			}

			if !status.Sealed {
				logger.Info(server, " is now unsealed.")
				*sealed = false // to ensure we only see this once for change of state from sealed
				break
			}
		}
	} else {
		s := *sealed
		if s {
			logger.Info(server, " is already unsealed.")
			*sealed = false // to ensure we only see this once for change of state from sealed
		}
	}
}

func monitorAndUnsealVaults(servers []string, unsealKeys []string, probeInterval int) {

	var sealed = make([]bool, len(servers)) // used to only report a vault server is 'unsealed' once (until it becomes sealed)

	for i, _ := range servers {
		sealed[i] = true // assume and indicate vault is sealed to start with
	}

	for {
		anySealed := false
		for i, server := range servers {
			if sealed[i] {
				anySealed = true
				// run serially to minimise logs
				checkAndUnsealVault(server, unsealKeys, &sealed[i])
				if sealed[i] {
					// its still sealed, so do specified delay
					time.Sleep(time.Duration(probeInterval) * time.Second)
				}
			}
		}
		if anySealed == false {
			break
		}
	}
}
