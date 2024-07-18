# GRPC Bi-Directional Remote User Store Sample (Go Lang)

This project contains a GRPC user store server, a GRPC remote agent and a sample user store manager.

For the demonstration purposes, the project only supports user authentication against a set of locally stored user credentials.

This involves running a bi-directional gRPC server and a client application that communicates with an on premise user store to perform user operations such as authentication.

The client should be kept private in a local environment and should only allow outbound communications to the server endpoint. The client initiates the connection with the server and creates a bi-directional communication channel.

The GRPC server exposes a GRPC endpoint to connect from third party applications such as the provided user store manager sample.

### How to run

1. navigate to the `userstore-server` directory and run the following command:
   ```bash
   go run .
   ```

2. navigate to the `agent` directory and run the following command:
   ```bash
   go run .
   ```

   Note: If you're using secrets in the `deployment.toml` file, you need to set the `SECRET_KEY` environment variable.

3. navigate to the `userstore-manager` directory and run the following command:

   ```bash
   go run user_store_manager.go
   ```

### Agent authentication

1. Start a MSSQL server database and create a database named `dbIdentityAgent`.
2. Run the SQL script `resources/db_script.sql` to create the required tables.
3. Configure the database connection variables in the `userstore-server/deployment.toml` file.

### Encrypting agent secrets

1. Add the `[secrets]` configuration at the bottom of the `agent/deployment.toml` file. Give an alias for the secret followed by the actual secret. Secret value should be enclosed in square brackets.
```toml
[secrets]
token = "[0ff93c70d1eb86972e1b9ac69cc8540bf8acf5a2021fe9dbfb621bb8c793c74c]"
```

2. Set `SECRET_KEY` environment variable to a 32-byte key. This key is used to encrypt and decrypt the secrets.
   ```bash
   export SECRET_KEY="51e6a32d699c43f7cbd7c62ba999c64a"
   ```

3. navigate to the `cipher-tool` directory and run the following command:
   ```bash
   go run . <path/to/the/deployment.toml/file>
   ```

4. Go back to the `deployment.toml` file and see that the secret values are encrypted.
   ```toml
   [secrets]
   token = "71ba5543e71db73d1e4e640d46601bc63820204a2e8017adb5f1ffca58dc1838ca873edb46b4ecb6ee7258c8cd4bc5acf59771f9220f0b5e328a7909ca9d3046b7856420821dd5dbfbf03d07b50530ee4f757b1ff21bfa9739bad816"
   ```

5. Now add the encrypted secrets to the relevant sections in the `deployment.toml` file by using a placeholder: `$secret{alias}`.
   ```toml
   [security]
   token = "$secret{token}"
   ```
