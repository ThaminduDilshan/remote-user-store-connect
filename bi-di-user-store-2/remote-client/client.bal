import ballerina/log;

// Server variables.
final string serverEndpoint = "http://localhost:9090";

// Client variables.
final string client_id = "client_001";
final string organization = "test_org_1";
final int noOfStreams = 2;

public function main() returns error? {
    
    log:printInfo("Starting the remote user store client...");

    GRPCServerClient ep = check new (serverEndpoint);

    // TODO: Need to do this in a loop to create multiple streams.
    // int i = 0;
    // while i < noOfStreams {
    //     log:printDebug("Creating a new stream with id: " + i.toString() + "...");
    //     i += 1;
    // }
    
    int streamId = 1;
    CommunicateStreamingClient streamingClient = check ep->communicate();

    // Start receiving remote requests in another strand.
    future<error?> f1 = start receiveServerRequests(streamingClient);

    // Send a client connection request to the server.
    RemoteMessage connectionRequest = {
        id: client_id,
        streamId: streamId.toString(),
        organization: organization,
        operationType: "CLIENT_CONNECT",
        data: {}
    };       
    check streamingClient->sendRemoteMessage(connectionRequest);

    // Wait for the receiveServerRequests function to complete.
    check wait f1;
}

// Function to keep receiving remote requests from the server.
function receiveServerRequests(CommunicateStreamingClient streamingClient) returns error? {

    RemoteMessage? remoteMessage = check streamingClient->receiveRemoteMessage();

    while !(remoteMessage is ()) {
        log:printInfo("Received a remote message with id: " + remoteMessage.id 
            + " for the operation type: " + remoteMessage.operationType);

        // Process the received message.
        error? result = handleUserStoreRequest(streamingClient, remoteMessage);

        if result is error {
            log:printError("Error occurred while processing the remote message with id: " + remoteMessage.id);
        }

        remoteMessage = check streamingClient->receiveRemoteMessage();
    }
}

// Function to process the user store request.
function handleUserStoreRequest(CommunicateStreamingClient streamingClient, RemoteMessage remoteMessage) 
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
                    streamId: remoteMessage.streamId,
                    organization: remoteMessage.organization,
                    operationType: "USERSTORE_OPERATION_RESPONSE",
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
                        streamId: remoteMessage.streamId,
                        organization: remoteMessage.organization,
                        operationType: USERSTORE_OPERATION_RESPONSE,
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
                    streamId: remoteMessage.streamId,
                    organization: remoteMessage.organization,
                    operationType: USERSTORE_OPERATION_RESPONSE,
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
                streamId: remoteMessage.streamId,
                organization: remoteMessage.organization,
                operationType: USERSTORE_OPERATION_RESPONSE,
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
