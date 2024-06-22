import ballerina/log;

// Server variables.
final string serverEndpoint = "http://localhost:9090";
final string client_id = "client_001";
final string organization = "test_org_1";

public function main() returns error? {

    log:printInfo("Starting the remote user store client...");

    RemoteUserStoreClient ep = check new (serverEndpoint);
    CommunicateStreamingClient streamingClient = check ep->communicate();

    // Start receiving remote requests in another strand.
    future<error?> f1 = start receiveRemoteRequests(streamingClient);

    // Send a client connection request to the server.
    RemoteMessage connectionRequest = {
        operationType: CLIENT_CONNECT,
        id: client_id,
        organization: organization
    };
    check streamingClient->sendRemoteMessage(connectionRequest);

    // Wait for the receiveRemoteRequests function to complete.
    check wait f1;
}

// Function to keep receiving remote requests and heart beats from the server.
// 
// + streamingClient - The streaming client to communicate with the server.
// + returns         - An error if an error occurs while receiving remote requests.
function receiveRemoteRequests(CommunicateStreamingClient streamingClient) returns error? {

    RemoteMessage? remoteMessage = check streamingClient->receiveRemoteMessage();

    while !(remoteMessage is ()) {

        if remoteMessage.operationType === SERVER_CONNECT {
            log:printInfo("Connection successful with the server: " + remoteMessage.id);
        } else if remoteMessage.operationType === SERVER_HEART_BEAT {
            log:printDebug("Received a heart beat from the server: " + remoteMessage.id);

            // Send a heart beat ack to the server.
            // Note: This is required to keep the bi-directional stream alive in both directions.
            RemoteMessage heartBeatAck = {
                operationType: CLIENT_HEART_BEAT_ACK,
                id: client_id
            };
            check streamingClient->sendRemoteMessage(heartBeatAck);
        } else {
            log:printInfo("Received a remote message with id: " + remoteMessage.id 
                + "for the operation type: " + remoteMessage.operationType);

            // Process the received message.
            error? processResult = processUserStoreRequest(streamingClient, remoteMessage);

            if processResult is error {
                log:printError("Error occurred while processing the remote message with id: " + remoteMessage.id);
            }
        }

        remoteMessage = check streamingClient->receiveRemoteMessage();
    }
}

// Function to process the user store request.
//
// + streamingClient - The streaming client to communicate with the server.
// + remoteMessage   - The remote message to process.
// + returns         - An error if an error occurs while processing the remote message.
function processUserStoreRequest(CommunicateStreamingClient streamingClient, RemoteMessage remoteMessage) 
        returns error? {

    string id = remoteMessage.id;
    
    match remoteMessage.operationType {
        DO_AUTHENTICATE => {
            json|error? authData = remoteMessage.data.toJson();

            if authData is error {
                log:printError("Error occurred while parsing the authentication data for the id: " + id);

                UserStoreError usError = {
                    _error: "INVALID_REQUEST",
                    message: "Unable to parse the authentication data."
                };
                
                RemoteMessage responseMessage = {
                    id: id,
                    operationType: USERSTORE_OPERATION_RESPONSE,
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
                        operationType: USERSTORE_OPERATION_RESPONSE,
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

                AuthenticationSuccessResponse|AuthenticationFailResponse authResponse = doAuthenticate(authRequest);

                RemoteMessage responseMessage = {
                    id: id,
                    operationType: USERSTORE_OPERATION_RESPONSE,
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
                operationType: USERSTORE_OPERATION_RESPONSE,
                organization: remoteMessage.organization,
                data: usError
            };
            check streamingClient->sendRemoteMessage(responseMessage);
        }
    }
}

// Function to authenticate the user.
//
// + authRequest - The authentication request.
// + returns     - The authentication response.
function doAuthenticate(AuthenticationRequest authRequest) 
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
            _error: "INVALID CREDENTIALS",
            message: "Provided username or password is incorrect. Please try again."
        };

        return response;
    }
}
