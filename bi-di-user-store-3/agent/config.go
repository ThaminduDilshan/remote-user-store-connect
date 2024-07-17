package main

import (
	"flag"
	"log"
	"os"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftoml"
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
}

func readConfig() Config {

	var config Config

	fs := flag.NewFlagSet("server_configs", flag.ExitOnError)

	fs.StringVar(&config.System.ID, "system.id", "agent", "Agent identifier")
	fs.StringVar(&config.System.Tenant, "system.tenant", "", "Tenant name")
	fs.StringVar(&config.System.UserStore, "system.user_store", "", "User store")
	fs.IntVar(&config.System.NoOfIdleConnections, "system.idle_connections", 10, "Number of idle connections")

	fs.StringVar(&config.Security.Token, "security.token", "", "Agent installation token")
	fs.StringVar(&config.HubService.Host, "hub_service.host", "localhost", "Hub service host")
	fs.IntVar(&config.HubService.Port, "hub_service.port", 9004, "Hub service port")

	// Parse the TOML file.
	err := ff.Parse(fs, os.Args[1:], ff.WithConfigFile("deployment.toml"), ff.WithConfigFileParser(fftoml.Parser))
	if err != nil {
		log.Fatal("Error reading the config file: ", err)
	}

	return config
}
