import ballerina/grpc;
import ballerina/protobuf;

public const string BIDI_STREAM_DESC = "0A11626964695F73747265616D2E70726F746F226F0A0B436861744D657373616765120E0A0269641801200128095202696412120A046E616D6518022001280952046E616D6512180A076D65737361676518032001280952076D65737361676512220A0C6F7267616E697A6174696F6E180420012809520C6F7267616E697A6174696F6E322E0A044368617412260A0463686174120C2E436861744D6573736167651A0C2E436861744D65737361676528013001620670726F746F33";

public isolated client class ChatClient {
    *grpc:AbstractClientEndpoint;

    private final grpc:Client grpcClient;

    public isolated function init(string url, *grpc:ClientConfiguration config) returns grpc:Error? {
        self.grpcClient = check new (url, config);
        check self.grpcClient.initStub(self, BIDI_STREAM_DESC);
    }

    isolated remote function chat() returns ChatStreamingClient|grpc:Error {
        grpc:StreamingClient sClient = check self.grpcClient->executeBidirectionalStreaming("Chat/chat");
        return new ChatStreamingClient(sClient);
    }
}

public client class ChatStreamingClient {
    private grpc:StreamingClient sClient;

    isolated function init(grpc:StreamingClient sClient) {
        self.sClient = sClient;
    }

    isolated remote function sendChatMessage(ChatMessage message) returns grpc:Error? {
        return self.sClient->send(message);
    }

    isolated remote function sendContextChatMessage(ContextChatMessage message) returns grpc:Error? {
        return self.sClient->send(message);
    }

    isolated remote function receiveChatMessage() returns ChatMessage|grpc:Error? {
        var response = check self.sClient->receive();
        if response is () {
            return response;
        } else {
            [anydata, map<string|string[]>] [payload, _] = response;
            return <ChatMessage>payload;
        }
    }

    isolated remote function receiveContextChatMessage() returns ContextChatMessage|grpc:Error? {
        var response = check self.sClient->receive();
        if response is () {
            return response;
        } else {
            [anydata, map<string|string[]>] [payload, headers] = response;
            return {content: <ChatMessage>payload, headers: headers};
        }
    }

    isolated remote function sendError(grpc:Error response) returns grpc:Error? {
        return self.sClient->sendError(response);
    }

    isolated remote function complete() returns grpc:Error? {
        return self.sClient->complete();
    }
}

public client class ChatChatMessageCaller {
    private grpc:Caller caller;

    public isolated function init(grpc:Caller caller) {
        self.caller = caller;
    }

    public isolated function getId() returns int {
        return self.caller.getId();
    }

    isolated remote function sendChatMessage(ChatMessage response) returns grpc:Error? {
        return self.caller->send(response);
    }

    isolated remote function sendContextChatMessage(ContextChatMessage response) returns grpc:Error? {
        return self.caller->send(response);
    }

    isolated remote function sendError(grpc:Error response) returns grpc:Error? {
        return self.caller->sendError(response);
    }

    isolated remote function complete() returns grpc:Error? {
        return self.caller->complete();
    }

    public isolated function isCancelled() returns boolean {
        return self.caller.isCancelled();
    }
}

public type ContextChatMessageStream record {|
    stream<ChatMessage, error?> content;
    map<string|string[]> headers;
|};

public type ContextChatMessage record {|
    ChatMessage content;
    map<string|string[]> headers;
|};

@protobuf:Descriptor {value: BIDI_STREAM_DESC}
public type ChatMessage record {|
    string id = "";
    string name = "";
    string message = "";
    string organization = "";
|};

