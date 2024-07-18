package config

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	System struct {
		ID                  string `toml:"id"`
		Tenant              string `toml:"tenant"`
		UserStore           string `toml:"user_store"`
		NoOfIdleConnections int    `toml:"idle_connections"`
	} `toml:"system"`

	Security struct {
		Token string `toml:"token"`
	} `toml:"security"`

	HubService struct {
		Host string `toml:"host"`
		Port int    `toml:"port"`
	} `toml:"hub_service"`

	Secrets map[string]string `toml:"secrets"`
}

func ReadConfig(filePath string) Config {

	var config Config

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Error opening config file: ", err)
	}
	defer file.Close()

	if _, err := toml.NewDecoder(file).Decode(&config); err != nil {
		log.Fatal("Error decoding config file: ", err)
	}

	return config
}
