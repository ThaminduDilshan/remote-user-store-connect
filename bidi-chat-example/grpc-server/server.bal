import ballerina/grpc;
import ballerina/io;
import ballerina/uuid;
import ballerina/time;


isolated map<RemoteJob[]> remoteJobs = {};
isolated map<RemoteResult> remoteResults = {};

isolated function addRemoteJob(string organization, RemoteJob job) {

    lock {
        if remoteJobs.hasKey(organization) {
            remoteJobs.get(organization).push(job.clone());
            io:println("Remote job added to the existing organization: ", organization, ". Job id: ", job.id);
        } else {
            RemoteJob[] orgJobs = [];
            orgJobs.push(job.clone());
            remoteJobs[organization] = orgJobs;
            io:println("Remote job added to the organization: ", organization, ". Job id: ", job.id);
        }
    }
}

isolated function getRemoteJob(string organization) returns RemoteJob|null {

    RemoteJob|null returnJob = null;

    lock {
        if remoteJobs.hasKey(organization) && remoteJobs.get(organization).length() > 0 {
            RemoteJob job = remoteJobs.get(organization).remove(0);

            io:println("Retrieving remote job from the organization: ", organization, ". Job id: ", job.id);

            returnJob = {
                id: job.id,
                organization: job.organization,
                name: job.name,
                message: job.message
            };
        }
    }

    return returnJob;
}

isolated function addRemoteResult(RemoteResult result) {

    lock {
        remoteResults[result.id] = result.clone();
        io:println("Remote result added with id: ", result.id);
    }
}

isolated function getRemoteResult(string id) returns RemoteResult|null {

    RemoteResult|null returnResult = null;

    lock {
        if remoteResults.hasKey(id) {
            returnResult = remoteResults.remove(id).clone();
        }
    }

    return returnResult;
}



@grpc:Descriptor {
    value: BIDI_STREAM_DESC
}
service "Chat" on new grpc:Listener(9090) {

    remote function chat(ChatChatMessageCaller caller, stream<ChatMessage, error?> clientStream) returns error? {

        future<error?> f1 = start sendOutMessage(caller);

        do {
            _ = check from ChatMessage message in clientStream
            do {
                io:println("Message received from client with id: ", message.id);

                if message.id !== "PING" {
                    io:println("Test");
                    RemoteResult result = {
                        id: message.id,
                        organization: message.organization,
                        status: message.name,
                        message: message.message                        
                    };

                    addRemoteResult(result);
                }

                // ChatMessage serverMessage = {
                //     name: "server",
                //     message: "PONG"
                // };
                // checkpanic caller->sendChatMessage(serverMessage);
            };
            check caller->complete();
        } on fail error err {
            io:println("Connection is closed with an error: ", err);
        }

        check wait f1;
    }

    remote isolated function externalCall(ServerRequest inMessage) returns ServerResponse {

        string id = uuid:createType1AsString();
        time:Seconds requestTimeout = 10;

        io:println("External call received. Assgined the id: " + id);

        RemoteJob job = {
            id: id,
            organization: inMessage.organization,
            name: inMessage.name,
            message: inMessage.message
        };

        addRemoteJob(inMessage.organization, job);

        // while: true, need to keep listening for a response.
        time:Utc startTime = time:utcNow();

        while (true) {
            time:Utc currentTime = time:utcNow();
            time:Seconds duration = time:utcDiffSeconds(currentTime, startTime);

            if duration > requestTimeout {
                io:println("Request timed out for the id: " + id);
                return {
                    status: "FAIL",
                    message: "Timeout out!"
                };
            }

            RemoteResult|null response = getRemoteResult(id);

            if response is RemoteResult {
                io:println("Found the response for the id: " + id);
                
                ServerResponse returnResponse = {
                    status: response.status,
                    message: response.message
                };

                return returnResponse;
            }
        }
    }
}

function sendOutMessage(ChatChatMessageCaller caller) returns error? {

    time:Seconds pingInterval = 10;
    time:Utc? lastPingTime = null;
    
    while true {
        RemoteJob? job = getRemoteJob("test");

        if (job == null) {
            // continue;

            if lastPingTime == null {
                serverPing(caller);
                lastPingTime = time:utcNow();
            } else {
                time:Utc currentTime = time:utcNow();
                time:Seconds duration = time:utcDiffSeconds(currentTime, lastPingTime);

                if duration > pingInterval {
                    serverPing(caller);
                    lastPingTime = time:utcNow();
                }
            }
        } else {
            io:println("Remote job received with id: " + job.id);

            ChatMessage message = {
                id: job.id,
                name: job.name,
                message: job.message,
                organization: job.organization
            };

            checkpanic caller->sendChatMessage(message);
        }

        // io:println("Remote job received with id: " + job.id);

        // ChatMessage message = {
        //     id: job.id,
        //     name: job.name,
        //     message: job.message,
        //     organization: job.organization
        // };

        // checkpanic caller->sendChatMessage(message);
    }
}

function serverPing(ChatChatMessageCaller caller) {

    ChatMessage message = {
        id: "PING",
        name: "server",
        message: "",
        organization: ""
    };

    checkpanic caller->sendChatMessage(message);
}
