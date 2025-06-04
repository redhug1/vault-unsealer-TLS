package main

import (
	"flag"
	"strings"

	logger "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	LogLevel                string   `json:"log_level"`
	Nodes                   []string `json:"nodes"`
	ProbeInterval           int      `json:"probe_interval"`
	UnsealKeys              []string `json:"unseal_keys"`
	VaultNomadServerToken   string   `json:"vault_nomad_server_token_id"`
	VaultToken              string   `json:"vault_token"`
	VaultConsulConnectToken string   `json:"vault_consul_connect_token_id"`
}

var configFilePath = flag.String("config-file-path", ".", "The path where config.json file to use with vault-unsealer is located")
var configFile = flag.String("config-file", "config", "The path where config.json file to use with vault-unsealer is located")

func newConfig() *Config {

	flag.Parse()

	conf := strings.TrimSuffix(*configFile, ".json")

	config := viper.New()
	replacer := strings.NewReplacer(".", "_")
	config.SetEnvKeyReplacer(replacer)
	config.AutomaticEnv()

	config.SetDefault("log.level", "info")
	config.SetDefault("nodes", []string{"http://localhost:8200"})
	config.SetDefault("unseal_threshold", 1)
	config.SetDefault("probe_interval", 10)
	config.SetDefault("unseal_keys", nil)

	config.AddConfigPath("config")
	config.SetConfigName(conf)   // Register config file name (no extension)
	config.SetConfigType("json") // Look for specific type
	config.AddConfigPath(*configFilePath)
	err := config.ReadInConfig()
	if err != nil {
		logger.Fatal(err)
	}

	return &Config{
		LogLevel:                config.GetString("log_level"),
		Nodes:                   config.GetStringSlice("nodes"),
		ProbeInterval:           config.GetInt("probe_interval"),
		UnsealKeys:              config.GetStringSlice("unseal_keys"),
		VaultNomadServerToken:   config.GetString("vault_nomad_server_token_id"),
		VaultToken:              config.GetString("vault_token"),
		VaultConsulConnectToken: config.GetString("vault_consul_connect_token_id"),
	}
}
