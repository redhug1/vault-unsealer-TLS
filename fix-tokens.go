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

func fixTokens(servers []string, vault_nomad_server_token_id, vault_token, vault_consul_connect_token_id string) {
	for _, server := range servers {
		fixToken("nomad-server", server, vault_nomad_server_token_id, vault_token)
	}

	// Generate consul connect vault token
	for _, server := range servers {
		fixToken("consul-connect", server, vault_consul_connect_token_id, vault_token)
	}
}

// Generate NNN vault token
func fixToken(token_name, server, vault_token_id, vault_token string) {

	logger.Info(server, " - checking token: "+token_name)

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

	// Check vault NNN token [ derived from same section name in ansible code by running it with -vvv ]
	postBody := "{\"token\":\"" + vault_token_id + "\"}"
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
	fmt.Println("Response from token lookup: " + token_name + " on server: " + server)
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, body, "", "  ")
	if err != nil {
		logger.Error("Error parsing JSON body:", err)
		return
	}
	fmt.Println(prettyJSON.String())

	if resp.Status == "200 OK" {
		logger.Info(server, " - token OK: "+token_name)
		return
	}

	logger.Info(server, " - token is NOT OK: "+token_name)

	// Create vault NNN token [ derived from same section name in ansible code by running it with -vvv ]
	logger.Info(server, " - token Creating: "+token_name)

	postBody = "{\"id\":\"" + vault_token_id + "\",\"period\":\"10m\",\"policies\":[\"" + token_name + "\"]}"
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
	fmt.Println("Response from token creation: " + token_name + " on server: " + server)

	err = json.Indent(&prettyJSON, body, "", "  ")
	if err != nil {
		logger.Error("Error parsing JSON body 2:", err)
		return
	}
	fmt.Println(prettyJSON.String())

	fmt.Println("response Status:", resp.Status)

	if resp.Status == "200 OK" {
		logger.Info(server, " - token Created OK: "+token_name)
		return
	}
	logger.Error(server, " - token ... Creation Failed ... invetigate and FIX !!: "+token_name)
}
