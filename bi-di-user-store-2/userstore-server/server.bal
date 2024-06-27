import ballerina/log;
import ballerina/grpc;
import ballerina/time;
import ballerina/uuid;

final string server_id = "server";

// Map to hold active client streams and callers based on the organization.
isolated map<RemoteClientData> remoteClients = {};

isolated function pushClientStream (string clientId, string organization, RemoteClientStreamData clientStream) {

    lock {
        if remoteClients.hasKey(organization) {
            RemoteClientData remoteClient = remoteClients.get(organization);
            remoteClient.streams[clientStream.streamId] = clientStream;
        } else {
            RemoteClientData remoteClient = {
                clientId: clientId,
                organization: organization,
                streams: {
                    id: clientStream
                }
            };
            remoteClients[organization] = remoteClient;
        }
    }
}

isolated function popClientStream (string organization) returns RemoteClientStreamData|null {

    RemoteClientStreamData|null streamToReturn = null;

    lock {
        if remoteClients.hasKey(organization) {
            RemoteClientData remoteClient = remoteClients.get(organization);
            string[] streamIds = remoteClient.streams.keys();

            if (streamIds.length() > 0) {
                return remoteClient.streams.remove(streamIds[0]);
            }
        }
    }

    return streamToReturn;
}

isolated function removeClientStream (string organization, string streamId) {

    lock {
        if remoteClients.hasKey(organization) {
            RemoteClientData remoteClient = remoteClients.get(organization);

            if remoteClient.streams.hasKey(streamId) {
                _ = remoteClient.streams.remove(streamId);
            }
        }
    }
}

@grpc:Descriptor {
    value: SERVER_DESC
}
service "GRPCServer" on new grpc:Listener(9090) {

    remote isolated function communicate(GRPCServerRemoteMessageCaller caller, 
                                stream<RemoteMessage, error?> clientStream) returns error? {

        // Organization of the client.
        string organization = "";
        string clientId = "";
        string streamId = "";

        do {
            _ = check from RemoteMessage message in clientStream
                do {
                    if message.operationType === "CLIENT_CONNECT" {
                        organization = message.organization;
                        clientId = message.id;
                        streamId = message.streamId;

                        RemoteClientStreamData remoteClientStream = {
                            clientId: clientId,
                            streamId: streamId,
                            caller: caller,
                            clientStream: clientStream
                        };

                        pushClientStream(clientId, organization, remoteClientStream);

                        log:printInfo("<client: " + clientId + ", stream: " + streamId 
                            + "> of organization: " + organization + " connected to the server.");
                        
                        RemoteMessage connectResponse = {
                            id: server_id,
                            streamId: streamId,
                            organization: organization,
                            operationType: "SERVER_CONNECT",
                            data: {}
                        };
                        checkpanic caller->sendRemoteMessage(connectResponse);
                        
                    } else {
                        log:printError("Invalid operation received from the client. Type: " + message.operationType);
                    }
                };

            check caller->complete();
            log:printInfo("<client: " + clientId + ", stream: " + streamId 
                            + "> of organization: " + organization + " disconnected from the server.");
            removeClientStream(organization, streamId);

        } on fail error err {
            log:printError("Connection for <client: " + clientId + ", stream: " + streamId 
                            + "> of organization: " + organization + " closed with error: " + err.message());
            removeClientStream(organization, streamId);
        }
    }
}

@grpc:Descriptor {
    value: SERVER_DESC
}
service "RemoteServer" on new grpc:Listener(9092) {

    remote isolated function authenticate(AuthenticationRequest authRequest) returns AuthenticationResponse {

        string id = uuid:createType1AsString();
        string organization = authRequest.organization;
        time:Seconds requestTimeout = 10;

        log:printInfo("Authentication request received for the organization: " 
            + authRequest.organization + ". Assigned the id: " + id);

        RemoteMessage remoteMessage = {
            id: id,
            streamId: "",
            organization: organization,
            operationType: DO_AUTHENTICATE,
            data: {
                "username": authRequest.username,
                "password": authRequest.password
            }
        };

        // Send message to the client and wait for the response.
        log:printInfo("Sending a remote message to the organization: " + organization);

        log:printDebug("Retrieving a client stream for the organization: " + organization);
        RemoteClientStreamData? remoteClientStreamData = popClientStream(organization);

        if remoteClientStreamData is null {
            log:printError("No active agent connection found for the organization: " + organization);

            return {
                username: authRequest.username,
                data: {
                    fields: {
                        status: {string_value: "ERROR"},
                        message: {string_value: "Authentication failed with the remote client."}
                    }
                }
            };
        }

        if remoteClientStreamData.caller is GRPCServerRemoteMessageCaller 
            && remoteClientStreamData.clientStream is stream<RemoteMessage> {
            
            // Send the message to the client.
            GRPCServerRemoteMessageCaller caller = remoteClientStreamData.caller;
            checkpanic caller->sendRemoteMessage(remoteMessage);

            time:Utc startTime = time:utcNow();
            
            // Wait for a response from the stream.
            while true {
                time:Utc currentTime = time:utcNow();
                time:Seconds elapsed = time:utcDiffSeconds(currentTime, startTime);

                if elapsed > requestTimeout {
                    log:printError("Request timed out for the id: " + id);

                    log:printDebug("Releasing the client stream of the organization: " + organization);
                    pushClientStream(remoteClientStreamData.clientId, organization, remoteClientStreamData);

                    return {
                        username: authRequest.username,
                        data: {
                            fields: {
                                status: {string_value: "TIMEOUT"},
                                message: {string_value: "Request timed out while waiting for a response from the client."}
                            }
                        }
                    };
                }

                var result = remoteClientStreamData.clientStream.next();

                if result is record {| RemoteMessage value; |} 
                    && result.value.operationType == USERSTORE_OPERATION_RESPONSE {
                    
                    RemoteMessage response = result.value;

                    log:printInfo("Remote response received for the id: " + id);

                    log:printDebug("Releasing the client stream of the organization: " + organization);
                    pushClientStream(remoteClientStreamData.clientId, organization, remoteClientStreamData);

                    return {
                        username: authRequest.username,
                        data: response.data
                    };
                } else if result is error {
                    log:printError("Request failed for the id: " + id + " with error: " + result.message());

                    log:printDebug("Releasing the client stream of the organization: " + organization);
                    pushClientStream(remoteClientStreamData.clientId, organization, remoteClientStreamData);

                    return {
                        username: authRequest.username,
                        data: {
                            fields: {
                                status: {string_value: "ERROR"},
                                message: {string_value: "Authentication request failed with the error: " + result.message()}
                            }
                        }
                    };
                }
            }
        } else {
            log:printError("Invalid client caller-stream pair found for the organization: " + organization);

            return {
                username: authRequest.username,
                data: {
                    fields: {
                        status: {string_value: "ERROR"},
                        message: {string_value: "Authentication failed with the remote client."}
                    }
                }
            };
        }
    }
}
