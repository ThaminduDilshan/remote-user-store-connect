# GRPC Bi-Directional Remote User Store Sample

This project contains a user store server and a remote client that can be operated in two modes.

1. Private Remote Client (GRPC)

For the demonstration purposes, the project only supports user authentication against a set of locally stored user credentials.

## Private Remote Client

This mode involves running a bi-directional gRPC server and a client application that communicates with an on premise user store to perform user operations such as authentication.

The client should be kept private in a local environment and should only allow outbound communications to the server endpoint. The client initiates the connection with the server and creates a bi-directional communication channel. Both parties will keep the communication alive by sending heart beat messages.

The server exposes a GRPC endpoint to connect from third party applications.

![Architecture diagram](resources/bidi-grpc-architecture.png)

### How to run

1. navigate to the `userstore-server` directory and run the following command:

   ```bash
   go run server.go
   ```
2. navigate to the `agent` directory and run the following command:

   ```bash
   go run agent.go
   ```
3. navigate to the `userstore-manager` directory and run the following command:

   ```bash
   go run user_store_manager.go
   ```