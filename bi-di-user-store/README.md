# GRPC Bi-Directional Remote User Store Sample

This project contains a sample bi-directional gRPC server and a client application that communicates with an on premise user store to perform user operations such as authentication.

For the demonstration purposes, the project only supports user authentication against a set of locally stored user credentials.

The client should be kept private in a local environment and should only allow outbound communications to the server endpoint. The client initiates the connection with the server and creates a bi-directional communication channel. Both parties will keep the communication alive by sending heart beat messages.

The server exposes a GRPC endpoint to connect from third party applications.

![Architecture diagram](resources/bidi-grpc-architecture.png)

## How to run

1. navigate to the `remote-us-server` directory and run the following command:
    ```bash
    bal run
    ```

2. navigate to the `grpc-client` directory and run the following command:
    ```bash
    bal run
    ```

3. Invoke the server with the following grpc command to authenticate a user:
    - Endpoint: grpc://localhost:9090
    - Method: RemoteUserStore/invokeUserStore
    - Payload:
        ```json
        {
            "organization": "test_org_1",
            "operationType": "DO_AUTHENTICATE",
            "data": {
                "fields": {
                    "username": {
                        "string_value": "user1"
                    },
                    "password": {
                        "string_value": "user1"
                    }
                }
            }
        }
        ```

    ![Postman request](resources/Screenshot%20from%202024-06-22%2014-35-36.png)

## Enable Debug Logs

Create a file named `Config.toml` in the directory where you'll be running the ballerina command with the following content:

```toml
[ballerina.log]
level = "DEBUG"
```
