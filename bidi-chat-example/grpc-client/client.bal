import ballerina/io;
// import ballerina/lang.runtime;

string[] allowed_names = ["test1", "test2", "test3"];

public function main() returns error? {

    ChatClient ep = check new ("http://localhost:9090");

    ChatStreamingClient streamingClient = check ep->chat();

    future<error?> f1 = start readInMessage(streamingClient);

    // while true {
    //     ChatMessage message = {
    //         id: "PING",
    //         name: "client_001",
    //         message: "",
    //         organization: "testorg"
    //     };
    //     check streamingClient->sendChatMessage(message);
    //     runtime:sleep(10);
    // }

    ChatMessage message = {
        id: "PING",
        name: "client_001",
        message: "",
        organization: "testorg"
    };
    check streamingClient->sendChatMessage(message);

    // check streamingClient->complete();
    
    check wait f1;
}

function readInMessage(ChatStreamingClient streamingClient) returns error? {

    ChatMessage? result = check streamingClient->receiveChatMessage();

    while !(result is ()) {
        if result.id !== "PING" {
            io:println("Received a chat message with id: " + result.id);
            io:println("Reading message => name: " + result.name + ", message: " + result.message);

            error? processMessageResult = checkpanic processMessage(streamingClient, result);
            if processMessageResult is error {
                io:println("Error processing message: " + processMessageResult.message());
            }
        } else {
            io:println("Received a PING message");

            ChatMessage message = {
                id: "PING",
                name: "client_001",
                message: "",
                organization: "testorg"
            };
            check streamingClient->sendChatMessage(message);
        }

        result = check streamingClient->receiveChatMessage();
    }
}

function processMessage(ChatStreamingClient streamingClient, ChatMessage message) returns error? {

    ChatMessage reply = {
        id: message.id,
        name: "",
        message: "",
        organization: message.organization
    };

    if allowed_names.indexOf(message.name) != (){
        reply.name = "Authorized";
        reply.message = "Hello " + message.name + "!";
    } else {
        reply.name = "Unauthorized";
        reply.message = "Hello " + message.name + "! You are not allowed to chat here.";
    }

    check streamingClient->sendChatMessage(reply);
}
