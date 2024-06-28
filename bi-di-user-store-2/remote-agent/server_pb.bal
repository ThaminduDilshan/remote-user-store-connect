import ballerina/grpc;
import ballerina/protobuf;

public const string SERVER_DESC = "0A0C7365727665722E70726F746F1A1C676F6F676C652F70726F746F6275662F7374727563742E70726F746F22B2010A0D52656D6F74654D657373616765120E0A02696418012001280952026964121A0A0873747265616D4964180220012809520873747265616D496412220A0C6F7267616E697A6174696F6E180320012809520C6F7267616E697A6174696F6E12240A0D6F7065726174696F6E54797065180420012809520D6F7065726174696F6E54797065122B0A046461746118052001280B32172E676F6F676C652E70726F746F6275662E537472756374520464617461323F0A0A4752504353657276657212310A0B636F6D6D756E6963617465120E2E52656D6F74654D6573736167651A0E2E52656D6F74654D65737361676528013001620670726F746F33";

public isolated client class GRPCServerClient {
    *grpc:AbstractClientEndpoint;

    private final grpc:Client grpcClient;

    public isolated function init(string url, *grpc:ClientConfiguration config) returns grpc:Error? {
        self.grpcClient = check new (url, config);
        check self.grpcClient.initStub(self, SERVER_DESC);
    }

    isolated remote function communicate() returns CommunicateStreamingClient|grpc:Error {
        grpc:StreamingClient sClient = check self.grpcClient->executeBidirectionalStreaming("GRPCServer/communicate");
        return new CommunicateStreamingClient(sClient);
    }
}

public client class CommunicateStreamingClient {
    private grpc:StreamingClient sClient;

    isolated function init(grpc:StreamingClient sClient) {
        self.sClient = sClient;
    }

    isolated remote function sendRemoteMessage(RemoteMessage message) returns grpc:Error? {
        return self.sClient->send(message);
    }

    isolated remote function sendContextRemoteMessage(ContextRemoteMessage message) returns grpc:Error? {
        return self.sClient->send(message);
    }

    isolated remote function receiveRemoteMessage() returns RemoteMessage|grpc:Error? {
        var response = check self.sClient->receive();
        if response is () {
            return response;
        } else {
            [anydata, map<string|string[]>] [payload, _] = response;
            return <RemoteMessage>payload;
        }
    }

    isolated remote function receiveContextRemoteMessage() returns ContextRemoteMessage|grpc:Error? {
        var response = check self.sClient->receive();
        if response is () {
            return response;
        } else {
            [anydata, map<string|string[]>] [payload, headers] = response;
            return {content: <RemoteMessage>payload, headers: headers};
        }
    }

    isolated remote function sendError(grpc:Error response) returns grpc:Error? {
        return self.sClient->sendError(response);
    }

    isolated remote function complete() returns grpc:Error? {
        return self.sClient->complete();
    }
}

public client class GRPCServerRemoteMessageCaller {
    private grpc:Caller caller;

    public isolated function init(grpc:Caller caller) {
        self.caller = caller;
    }

    public isolated function getId() returns int {
        return self.caller.getId();
    }

    isolated remote function sendRemoteMessage(RemoteMessage response) returns grpc:Error? {
        return self.caller->send(response);
    }

    isolated remote function sendContextRemoteMessage(ContextRemoteMessage response) returns grpc:Error? {
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

public type ContextRemoteMessageStream record {|
    stream<RemoteMessage, error?> content;
    map<string|string[]> headers;
|};

public type ContextRemoteMessage record {|
    RemoteMessage content;
    map<string|string[]> headers;
|};

@protobuf:Descriptor {value: SERVER_DESC}
public type RemoteMessage record {|
    string id = "";
    string streamId = "";
    string organization = "";
    string operationType = "";
    map<anydata> data = {};
|};

