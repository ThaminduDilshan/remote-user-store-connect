import ballerina/grpc;
import ballerina/time;
import ballerina/log;
import ballerina/uuid;
import ballerina/io;

string server_id = "server";

@grpc:Descriptor {
    value: REMOTE_US_DESC
}
service "RemoteUserStore" on new grpc:Listener(9090) {

    remote function communicate(RemoteUserStoreRemoteMessageCaller caller, 
                                stream<RemoteMessage, error?> clientStream) returns error? {

        // Organization of the client.
        string organization = "testorg";
        
        // Start sending remote requests in another strand.
        // TODO: Need to check why this is not getting called.
        future<error?> f1 = start sendRemoteMessages(caller, organization);

        do {
            _ = check from RemoteMessage message in clientStream
                do {
                    if message.operationType === CLIENT_CONNECT {
                        log:printInfo("Client connected to the server.");
                    } else if message.operationType === SERVER_HEART_BEAT_ACK {
                        log:printInfo("Heart beat ack received from the client.");
                    } else if message.operationType === USERSTORE_OPERATION_RESPONSE {
                        log:printInfo("Remote response received from the client with id: " + message.id);

                        RemoteResponse remoteResponse = {
                            id: message.id,
                            data: message.data
                        };

                        addRemoteResponse(remoteResponse);
                    } else {
                        log:printError("Invalid operation received from the client. Type: " + message.operationType);
                    }
                };

            check caller->complete();
            log:printInfo("Client disconnected from the server.");
        } on fail error err {
            log:printError("Connection closed with the error: " + err.message());
        }

        // Wait for the remote message sender to complete.
        check wait f1;
    }

    remote isolated function invokeUserStore(UserStoreRequest usRequest) returns UserStoreResponse {

        string id = uuid:createType1AsString();
        time:Seconds requestTimeout = 10;

        log:printInfo("User store request received with the type: " + usRequest.operationType 
            + ". Assigned the id: " + id);

        RemoteJob job = {
            id: id,
            operationType: usRequest.operationType,
            organization: usRequest.organization,
            data: usRequest.data
        };

        addRemoteJob(usRequest.organization, job);
        log:printInfo("Remote job added for the id: " + id);

        // Wait for the response.
        time:Utc requestStartTime = time:utcNow();

        while true {
            time:Utc currentTime = time:utcNow();
            time:Seconds duration = time:utcDiffSeconds(currentTime, requestStartTime);

            if duration > requestTimeout {
                log:printError("Request timed out for the id: " + id);

                return {
                    operationType: usRequest.operationType,
                    organization: usRequest.organization,
                    data: {
                        status: "TIMEOUT",
                        message: "Request timed out while waiting for a response from the client."
                    }
                };
            }

            RemoteResponse? response = getRemoteResponse(id);

            if response is RemoteResponse {
                log:printInfo("Remote response received for the id: " + id);

                return {
                    operationType: usRequest.operationType,
                    organization: usRequest.organization,
                    data: response.data
                };
            }
        }
    }
}

// Send remote messages to the client.
//
// + caller         - Remote message caller.
// + organization   - Organization.
// + return - Error if there is an error in sending the remote messages.
function sendRemoteMessages(RemoteUserStoreRemoteMessageCaller caller, string organization) returns error? {

    io:println("Remote message sender started");

    time:Seconds heartBeatInterval = 10;
    time:Utc? lastHeartBeatTime = null;

    while true {
        RemoteJob? job = getRemoteJob(organization);

        if job == null {
            // If no job are there, send heart beat messages to the cient to keep the stream alive.
            if lastHeartBeatTime == null {
                serverHeartBeat(caller);
                lastHeartBeatTime = time:utcNow();
            } else {
                time:Utc currentTime = time:utcNow();
                time:Seconds duration = time:utcDiffSeconds(currentTime, lastHeartBeatTime);

                if duration > heartBeatInterval {
                    serverHeartBeat(caller);
                    lastHeartBeatTime = time:utcNow();
                }
            }
        } else {
            log:printInfo("Remote job retrieve with id: " + job.id + " for the organization: " + organization);

            RemoteMessage remoteMessage = {
                id: job.id,
                operationType: job.operationType,
                organization: job.organization,
                data: job.data
            };

            checkpanic caller->sendRemoteMessage(remoteMessage);
        }
    }
}

// Send a heart beat message to the client.
// 
// + caller - Remote message caller.
function serverHeartBeat(RemoteUserStoreRemoteMessageCaller caller) {
    
    RemoteMessage heartBeatMessage = {
        operationType: SERVER_HEART_BEAT,
        id: server_id
    };

    log:printDebug("Sending heart beat message to the client");

    checkpanic caller->sendRemoteMessage(heartBeatMessage);
}
