package main

import (
	"bi-di-user-store-3/agent/config"
	"bi-di-user-store-3/agent/sec_vault"
	"log"
	"os"
	"strings"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatalf("Please provide the configuration file path.")
	}

	// Read the first argument
	arg := os.Args[1]

	// Read configurations from the deployment.toml file.
	log.Println("Reading configurations.")
	config := config.ReadConfig(arg)

	// Read the AES key from environment variables.
	encSecretKey, keyExists := os.LookupEnv("SECRET_KEY")
	if !keyExists && encSecretKey == "" {
		log.Fatalf("Secret key not found in the environment variables.")
	}

	// Read the toml file as text.
	fileContent, err := os.ReadFile(arg)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
	fileContentStr := string(fileContent)

	for _, secretValue := range config.Secrets {
		encryptedSecret, err := sec_vault.Encrypt(secretValue, []byte(encSecretKey))
		if err != nil {
			log.Fatalf("Error encrypting secret key: %v", err)
		}

		log.Printf("Encrypting secret key: %s", secretValue)

		// Replace the secret value with the encrypted value in the file content.
		fileContentStr = strings.ReplaceAll(fileContentStr, secretValue, encryptedSecret)
	}

	// Write the updated file content to the deployment.toml file.
	err = os.WriteFile(arg, []byte(fileContentStr), 0644)
	if err != nil {
		log.Fatalf("Error updating the config file: %v", err)
	}

	log.Println("Secret values encrypted successfully.")
}
