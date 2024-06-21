import ballerina/log;
import ballerina/grpc;

RemoteBiDiServiceClient ep = check new ("http://localhost:9090");

public function main() returns error? {
    
    log:printInfo("Starting the client...");

    log:printInfo("Connecting to the server...");
    CommunicateStreamingClient streamingClient = check ep->communicate();

    // Register caller in the server.
    RemoteResponse connectRequest = {
        uuid: "",
        status: "CONNECT",
        message: "",
        data: {
            organization: "org1"
        }
    };
    grpc:Error? sendRemoteResponse = streamingClient->sendRemoteResponse(connectRequest);
    if sendRemoteResponse is grpc:Error {
        log:printError("Error occurred while sending the CONNECT request.", sendRemoteResponse);
        return;
    }

    future<error?> f = start readRequest(streamingClient);
    log:printInfo("Waiting for the server communication...");

    // check streamingClient->complete();
    // log:printInfo("Client completed the communication...");
    
    check wait f;

    log:printInfo("Client stopped...");

    // do {
    //     _ = check from RemoteRequest request in streamingClient
    //         do {
    //             string contextId = request.uuid;
    //             log:printInfo("Received a request with context ID: ", contextId);

    //             RemoteResponse response = processRequest(request);
    //             checkpanic streamingClient->sendRemoteResponse(response);
    //         }
    // } on fail error err {
    //     log:printError("The connection is closed with an error: ", err);
    // }
}

function readRequest(CommunicateStreamingClient streamingClient) returns error? {

    RemoteRequest|grpc:Error? request = streamingClient->receiveRemoteRequest();

    if (request is RemoteRequest) {
        string contextId = request.uuid;
        log:printInfo("Received a request with context ID: " + contextId);

        RemoteResponse|null response = processRequest(request);

        if response is RemoteResponse {
            checkpanic streamingClient->sendRemoteResponse(response);
        }
    } else {
        log:printError("Error occurred while receiving the request.", request);
    }
}


function processRequest(RemoteRequest request) returns RemoteResponse|null {
    
    string _type = request.'type;
    string contextId = request.uuid;

    match _type {
        "CONNECTED" => {
            log:printInfo("Successfully connected to the server.");
        }
        "AUTHENTICATE" => {
            log:printInfo("Processing an AUTHENTICATE request with context ID: " + contextId);

            return processAuthenticationRequest(request);
        }
        _ => {
            log:printError("Invalid request type: " + _type);
            
            return {
                uuid: contextId,
                status: "ERROR",
                message: "Invalid request type",
                data: {}
            };
        }
    }
};

function processAuthenticationRequest(RemoteRequest request) returns RemoteResponse {
    
    string contextId = request.uuid;

    log:printInfo("Processing an AUTHENTICATE request with context ID: " + contextId);

    json requestData = request.data.toJson();
    json|error username = requestData.username;
    json|error password = requestData.password;

    if username is error || password is error {
        log:printError("Invalid request data for AUTHENTICATE request with context ID: " + contextId);
        
        return {
            uuid: contextId,
            status: "FAIL",
            message: "Invalid credentials",
            data: {}
        };
    }

    string usernameStr = username.toString();
    string passwordStr = password.toString();

    // TODO: Simple logic to return user authentication status.
    if usernameStr == "test1" && passwordStr == "test1" {
        log:printInfo("User authenticated successfully with context ID: " + contextId);
        
        return {
            uuid: contextId,
            status: "SUCCESS",
            message: "User authenticated successfully",
            data: {
                userId: 1,
                username: usernameStr,
                email: "user1@wso2.com"
            }
        };
    } else {
        log:printInfo("User authentication failed with context ID: " + contextId);
        
        return {
            uuid: contextId,
            status: "FAIL",
            message: "Invalid credentials",
            data: {}
        };
    }
}