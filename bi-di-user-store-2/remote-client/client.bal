import ballerina/log;

// Server variables.
final string serverEndpoint = "http://localhost:9090";

// Client variables.
final string client_id = "client_001";
final string organization = "test_org_1";

configurable string clientMode = "PRIVATE";

public function main() returns error? {
    
    log:printInfo("Starting the remote user store client...");

    // Starting the client in private mode. I.e. a GRPC client.
    log:printInfo("Client mode is set to: PRIVATE");

    RemoteUserStoreClient ep = check new (serverEndpoint);
    CommunicateStreamingClient streamingClient = check ep->communicate();

    // Start receiving remote requests in another strand.
    future<error?> f1 = start handleReceiveRemoteRequests(streamingClient);

    // Send a client connection request to the server.
    RemoteMessage connectionRequest = {
        operationType: "CLIENT_CONNECT",
        id: client_id,
        organization: organization,
        data: {}
    };       
    check streamingClient->sendRemoteMessage(connectionRequest);

    // Wait for the handleReceiveRemoteRequests function to complete.
    check wait f1;
}

// Function to keep receiving remote requests from the server.
function handleReceiveRemoteRequests(CommunicateStreamingClient streamingClient) returns error? {

    RemoteMessage? remoteMessage = check streamingClient->receiveRemoteMessage();

    while !(remoteMessage is ()) {
        log:printInfo("Received a remote message with id: " + remoteMessage.id 
            + " for the operation type: " + remoteMessage.operationType);

        // Process the received message.
        error? processResult = handleProcessUserStoreRequest(streamingClient, remoteMessage);

        if processResult is error {
            log:printError("Error occurred while processing the remote message with id: " + remoteMessage.id);
        }

        remoteMessage = check streamingClient->receiveRemoteMessage();
    }
}

// Function to process the user store request.
function handleProcessUserStoreRequest(CommunicateStreamingClient streamingClient, RemoteMessage remoteMessage) 
        returns error? {

    string id = remoteMessage.id;
    
    match remoteMessage.operationType {
        "DO_AUTHENTICATE" => {
            json|error? authData = remoteMessage.data.toJson();

            if authData is error {
                log:printError("Error occurred while parsing the authentication data for the id: " + id);

                UserStoreError usError = {
                    _error: "INVALID_REQUEST",
                    message: "Unable to parse the authentication data."
                };
                
                RemoteMessage responseMessage = {
                    id: id,
                    operationType: "USERSTORE_OPERATION_RESPONSE",
                    organization: remoteMessage.organization,
                    data: usError
                };
                check streamingClient->sendRemoteMessage(responseMessage);
                return;
            } else {
                json|error? userName = authData.username;
                json|error? password = authData.password;

                if userName is error || password is error {
                    log:printError("Error occurred while parsing username/ password for the id: " + id);
                    
                    UserStoreError usError = {
                        _error: "INVALID_REQUEST",
                        message: "Unable to parse the authentication data."
                    };
                    
                    RemoteMessage responseMessage = {
                        id: id,
                        operationType: "USERSTORE_OPERATION_RESPONSE",
                        organization: remoteMessage.organization,
                        data: usError
                    };
                    check streamingClient->sendRemoteMessage(responseMessage);
                    return;
                }

                string userNameStr = check userName.ensureType(string);
                string passwordStr = check password.ensureType(string);

                AuthenticationRequest authRequest = {
                    username: userNameStr,
                    password: passwordStr
                };

                AuthenticationSuccessResponse|AuthenticationFailResponse authResponse = authenticateUser(authRequest);

                RemoteMessage responseMessage = {
                    id: id,
                    operationType: "USERSTORE_OPERATION_RESPONSE",
                    organization: remoteMessage.organization,
                    data: authResponse
                };
                check streamingClient->sendRemoteMessage(responseMessage);
                return;
            }
        }
        _ => {
            log:printError("Invalid operation type received for the id: " + id);
            UserStoreError usError = {
                _error: "INVALID_REQUEST",
                message: "Invalid operation type received."
            };
            
            RemoteMessage responseMessage = {
                id: id,
                operationType: "USERSTORE_OPERATION_RESPONSE",
                organization: remoteMessage.organization,
                data: usError
            };
            check streamingClient->sendRemoteMessage(responseMessage);
        }
    }
}

// Function to authenticate the user.
function authenticateUser(AuthenticationRequest authRequest) 
        returns AuthenticationSuccessResponse|AuthenticationFailResponse {

    if USER_CREDENTIALS.hasKey(authRequest.username) 
            && USER_CREDENTIALS.get(authRequest.username) == authRequest.password {
        
        User user = USER_DATA.get(authRequest.username);

        AuthenticationSuccessResponse response = {
            username: authRequest.username,
            userId: user.userId,
            email: user.email
        };

        return response;
    } else {
        AuthenticationFailResponse response = {
            username: authRequest.username,
            _error: "INVALID_CREDENTIALS",
            message: "Provided username or password is incorrect. Please try again."
        };

        return response;
    }
}
