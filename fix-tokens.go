package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	logger "github.com/sirupsen/logrus"
)

func fixTokens(servers []string, vault_nomad_server_token string, vault_token string) {
	for _, server := range servers {
		nomadToken(server, vault_nomad_server_token, vault_token)
		break //!!! trash once above function call does all thats needed
	}

	// Generate consul connect vault token
	/*
		for _, server := range servers {
			consulConnectToken(server, unsealKeys)
			//   Check vault consul-connect token

			//   Create vault consul-connect token
		}
	*/
}

// Generate nomad server vault token
func nomadToken(server string, vault_nomad_server_token string, vault_token string) {

	logger.Info(server, " - checking nomad token ...")

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

	// Check vault nomad-server [ derived from same section name in ansible code by running it with -vvv ]
	postBody := "{\"token\":\"" + vault_nomad_server_token + "\"}"
	jsonStr := []byte(postBody)
	req, err := http.NewRequest("POST", server+"/v1/auth/token/lookup", bytes.NewBuffer(jsonStr))
	if err != nil {
		logger.Error("Error setting up NewRequest for POST:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", vault_token)
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error doing POST for token lookup:", err)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading response body:", err)
		return
	}

	// show response in formatted JSON
	//!!! comment out showing the response once code working - beacuse its showing sensitive values
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, body, "", "  ")
	if err != nil {
		logger.Error("Error parsing JSON body:", err)
		return
	}
	fmt.Println(prettyJSON.String())

	if resp.Status == "200 OK" {
		logger.Info(server, " - nomad token ... is OK")
		return
	}

	logger.Info(server, " - nomad token ... is NOT OK")

	// Create vault nomad-server [ derived from same section name in ansible code by running it with -vvv ]
	logger.Info(server, " - nomad token ... Creating ...")

	postBody = "{\"id\":\"" + vault_nomad_server_token + "\",\"period\":\"10m\",\"policies\":[\"nomad-server\"]}"
	jsonStr = []byte(postBody)
	req, err = http.NewRequest("POST", server+"/v1/auth/token/create-orphan", bytes.NewBuffer(jsonStr))
	if err != nil {
		logger.Error("Error setting up NewRequest for POST 2:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", vault_token)
	resp, err = client.Do(req)
	if err != nil {
		logger.Error("Error doing POST for token lookup 2:", err)
		return
	}

	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading response body 2:", err)
		return
	}

	// show response in formatted JSON
	//!!! comment out showing the response once code working - beacuse its showing sensitive values
	err = json.Indent(&prettyJSON, body, "", "  ")
	if err != nil {
		logger.Error("Error parsing JSON body 2:", err)
		return
	}
	fmt.Println(prettyJSON.String())

	fmt.Println("response Status:", resp.Status)

	if resp.Status == "200 OK" {
		logger.Info(server, " - nomad token ... Created OK")
		return
	}
	logger.Error(server, " - nomad token ... Creation Failed ... invetigate and FIX !!!")
}
