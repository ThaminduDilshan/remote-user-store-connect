import ballerina/grpc;
import ballerina/log;
import ballerina/uuid;
import ballerina/time;

time:Seconds requestTimeout = 10;

// TODO: isolate?
map<RemoteBiDiServiceRemoteRequestCaller> remoteRequestCallers = {};

function addCaller(string key, RemoteBiDiServiceRemoteRequestCaller caller) {
    
    if (remoteRequestCallers[key] == null) {
        remoteRequestCallers[key] = caller;
        log:printInfo("Caller added for key: " + key);
    }
}

function getCaller(string key) returns RemoteBiDiServiceRemoteRequestCaller|null {

    log:printInfo("Caller retrieved for key: " + key);
    return remoteRequestCallers[key];
}

// TODO: isolate.
map<RemoteResponse> remoteResponses = {};

function addResponse(string key, RemoteResponse response) {
    
    if (remoteResponses[key] == null) {
        remoteResponses[key] = response;
        log:printInfo("Response added for key: " + key);
    } else {
        log:printError("A response already exists for the key: " + key);
    }
}

function getResponse(string key) returns RemoteResponse|null {

    if remoteResponses.hasKey(key) {
        log:printInfo("Response retrieved for key: " + key);
        return remoteResponses.remove(key);
    }

    return null;
}


@grpc:Descriptor {
    value: REMOTE_US_DESC
}
service "RemoteBiDiService" on new grpc:Listener(9090) {
    
    remote function communicate(RemoteBiDiServiceRemoteRequestCaller caller, stream<RemoteResponse, error?> clientStream) {

        do {
            _ = check from RemoteResponse response in clientStream
                do {
                    log:printInfo("Received a message from client");
                    string status = response.status;

                    if (status == "CONNECT") {
                        // TODO: Read org.
                        addCaller("org1", caller);
                        
                        // // Send ack.
                        // RemoteRequest connectResponse = {
                        //     uuid: "",
                        //     'type: "CONNECTED",
                        //     data: {}
                        // };

                        // grpc:Error? sendRemoteRequest = caller->sendRemoteRequest(connectResponse);
                        // if sendRemoteRequest is grpc:Error {
                        //     log:printWarn("Error occurred while sending the connection response.", sendRemoteRequest);
                        // }
                    } else {
                        addResponse(response.uuid, response);
                    }
                };
        } on fail error err {
            log:printError("The connection is closed with an error.", err);
        }
    }

    // TODO: There needs to be another function to be called from outside (i.e. From Postman/ UserStoreManager).
    remote function authenticate(AuthenticationRequest authRequest) returns AuthenticationResponse {

        string organization = authRequest.organization;
        log:printInfo("Received an authentication request for the organization: " + organization);
        
        RemoteBiDiServiceRemoteRequestCaller|null caller = getCaller(organization);
        
        if caller is null {
            log:printInfo("Caller is not found for the organization: " + organization);
            return {
                status: "FAIL",
                message: "Caller is not found for the organization: " + organization,
                data: {}
            };
        }

        //
        string contextId = uuid:createType1AsString();
        RemoteRequest remoteRequest = {
            uuid: contextId,
            'type: "AUTHENTICATE",
            data: {
                username: authRequest.username,
                password: authRequest.password
            }
        };

        // Send the request to the caller.
        error? e = caller->sendRemoteRequest(remoteRequest);
        if e is error {
            log:printError("Error occurred while sending the request.", e);
            return {
                status: "FAIL",
                message: "Error occurred while sending the request.",
                data: {}
            };
        }

        // // TODO: Test
        // grpc:Error? complete = caller->complete();
        // if complete is grpc:Error {
        //     log:printError("Error occurred while completing the call.", complete);
        // }

        // while: true, need to keep listening for a response.
        time:Utc startTime = time:utcNow();
        
        while (true) {
            time:Utc currentTime = time:utcNow();
            time:Seconds duration = time:utcDiffSeconds(currentTime, startTime);

            if duration > requestTimeout {
                log:printError("Request timed out for the contextId: " + contextId);
                return {
                    status: "FAIL",
                    message: "Request timed out.",
                    data: {}
                };
            }

            RemoteResponse|null response = getResponse(contextId);
            if response is RemoteResponse {
                log:printInfo("Found a response for the contextId: " + contextId);
                
                AuthenticationResponse authResponse = {
                    status: response.status,
                    message: response.message,
                    data: response.data
                };

                return authResponse;
            }

            // // Sleep for a while.
            // runtime:sleep(1);
        }
    }
}
