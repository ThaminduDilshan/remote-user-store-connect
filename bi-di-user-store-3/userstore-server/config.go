package main

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server struct {
		Id   string `toml:"id"`
		Host string `toml:"host"`
		Port int    `toml:"port"`
	} `toml:"server"`

	ConnectionParam struct {
		MaxMessageSize int `toml:"max_message_size"`
		RequestTimeout int `toml:"request_timeout"`
	} `toml:"connection_param"`

	Database struct {
		Host     string `toml:"host"`
		Port     int    `toml:"port"`
		User     string `toml:"user"`
		Password string `toml:"password"`
		Name     string `toml:"name"`
	} `toml:"database"`
}

func readConfig() Config {

	var config Config

	file, err := os.Open("deployment.toml")
	if err != nil {
		log.Fatal("Error opening config file: ", err)
	}
	defer file.Close()

	if _, err := toml.NewDecoder(file).Decode(&config); err != nil {
		log.Fatal("Error decoding config file: ", err)
	}

	return config
}
