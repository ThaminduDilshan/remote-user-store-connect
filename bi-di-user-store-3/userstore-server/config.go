package main

import (
	"flag"
	"log"
	"os"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftoml"
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

	fs := flag.NewFlagSet("server_configs", flag.ExitOnError)

	fs.StringVar(&config.Server.Id, "server.id", "server", "Server identifier")
	fs.StringVar(&config.Server.Host, "server.host", "localhost", "Server host")
	fs.IntVar(&config.Server.Port, "server.port", 9004, "Server port")
	fs.IntVar(&config.ConnectionParam.MaxMessageSize, "connection_param.max_message_size", 104857600, "Max message size")
	fs.IntVar(&config.ConnectionParam.RequestTimeout, "connection_param.request_timeout", 15, "Request timeout")

	// Read database configurations.
	fs.StringVar(&config.Database.Host, "database.host", "localhost", "Database host")
	fs.IntVar(&config.Database.Port, "database.port", 1433, "Database port")
	fs.StringVar(&config.Database.User, "database.user", "user", "Database user")
	fs.StringVar(&config.Database.Password, "database.password", "password", "Database password")
	fs.StringVar(&config.Database.Name, "database.name", "db", "Database name")

	// Parse the TOML file.
	err := ff.Parse(fs, os.Args[1:], ff.WithConfigFile("deployment.toml"), ff.WithConfigFileParser(fftoml.Parser))
	if err != nil {
		log.Fatal("Error reading the config file: ", err)
	}

	return config
}
