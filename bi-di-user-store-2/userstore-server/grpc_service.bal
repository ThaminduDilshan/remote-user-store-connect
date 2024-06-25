import ballerina/log;
import ballerina/grpc;
import ballerina/time;
import ballerina/uuid;

final string server_id = "server";

// Maps to hold active client streams and callers based on the organization
map<RemoteUserStoreRemoteMessageCaller> callers = {};
map<stream<RemoteMessage, error?>> streams = {};

@grpc:Descriptor {
    value: REMOTE_US_DESC
}
service "RemoteUserStore" on new grpc:Listener(9090) {

    remote function communicate(RemoteUserStoreRemoteMessageCaller caller, 
                                stream<RemoteMessage, error?> clientStream) returns error? {

        // Organization of the client.
        string organization = "";

        do {
            _ = check from RemoteMessage message in clientStream
                do {
                    if message.operationType === "CLIENT_CONNECT" {
                        organization = message.organization;

                        log:printInfo("Client: " + message.id + " of organization: " + organization 
                            + " connected to the server.");

                        lock {
                            callers[organization] = caller;
                            streams[organization] = clientStream;
                        }
                        RemoteMessage connectResponse = {
                            operationType: "SERVER_CONNECT",
                            id: server_id,
                            organization: organization,
                            data: {}
                        };
                        checkpanic caller->sendRemoteMessage(connectResponse);
                        
                    } else {
                        log:printError("Invalid operation received from the client. Type: " + message.operationType);
                    }
                };

            check caller->complete();
            log:printInfo("Client disconnected from the server.");
            lock {
                _ = callers.remove(organization);
                _ = streams.remove(organization);
            }
        } on fail error err {
            log:printError("Connection closed with the error: " + err.message());
            lock {
                _ = callers.remove(organization);
                _ = streams.remove(organization);
            }
        }
    }
}

@grpc:Descriptor {
    value: REMOTE_US_DESC
}
service "RemoteServer" on new grpc:Listener(9092) {

    remote function invokeUserStore(UserStoreRequest usRequest) returns UserStoreResponse {

        log:printInfo("User store request received with the type: " + usRequest.operationType 
            + ". Assigned the id: " + usRequest.organization);

        string id = uuid:createType1AsString();
        time:Seconds requestTimeout = 10;

        RemoteMessage job = {
            id: id,
            operationType: usRequest.operationType,
            organization: usRequest.organization,
            data: usRequest.data
        };

        log:printInfo("Remote job added for the id: " + id);

        // Send job to the client and wait for the response
        return sendAndReceiveJob(usRequest.organization, job, requestTimeout);
    }
}

// Function to send a job to the client and wait for the response.
function sendAndReceiveJob(string organization, RemoteMessage job, time:Seconds timeout) returns UserStoreResponse {

    log:printInfo("Sending job to organization: " + organization);
    RemoteUserStoreRemoteMessageCaller? clientCaller = callers[organization];
    stream<RemoteMessage, error?>? clientStream = streams[organization];

    if clientCaller is RemoteUserStoreRemoteMessageCaller && clientStream is stream<RemoteMessage, error?> {
        checkpanic clientCaller->sendRemoteMessage(job);

        // Wait for the response directly in the stream
        time:Utc startTime = time:utcNow();

        while true {
            time:Utc currentTime = time:utcNow();
            time:Seconds elapsed = time:utcDiffSeconds(currentTime, startTime);

            if elapsed > timeout {
                log:printError("Request timed out for the id: " + job.id);
                return {
                    operationType: job.operationType,
                    organization: job.organization,
                    data: {
                        fields: {
                            status: {string_value: "TIMEOUT"},
                            message: {string_value: "Request timed out while waiting for a response from the client."}
                        }
                    }
                };
            }

            var result = clientStream.next();
            if result is record {| RemoteMessage value; |} {
                RemoteMessage response = result.value;
                if response.id == job.id {
                    log:printInfo("Remote response received for the id: " + job.id);
                    return {
                        operationType: job.operationType,
                        organization: job.organization,
                        data: response.data
                    };
                }
            } else if result is error {
                log:printError("Request failed for the id: " + job.id + " with error: " + result.message());
                return {
                    operationType: job.operationType,
                    organization: job.organization,
                    data: {
                        fields: {
                            status: {string_value: "ERROR"},
                            message: {string_value: result.message()}
                        }
                    }
                };
            }
        }
    } else {
        log:printError("No active agent connection for organization: " + organization);
        return {
            operationType: job.operationType,
            organization: job.organization,
            data: {
                fields: {
                    status: {string_value: "ERROR"},
                    message: {string_value: "No active agent connection for organization: " + organization}
                }
            }
        };
    }
}
