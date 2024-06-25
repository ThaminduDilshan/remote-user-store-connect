import ballerina/grpc;
import ballerina/protobuf;

public const string REMOTE_US_DESC = "0A0F72656D6F74655F75732E70726F746F1A1C676F6F676C652F70726F746F6275662F7374727563742E70726F746F2296010A0D52656D6F74654D657373616765120E0A0269641801200128095202696412240A0D6F7065726174696F6E54797065180220012809520D6F7065726174696F6E5479706512220A0C6F7267616E697A6174696F6E180320012809520C6F7267616E697A6174696F6E122B0A046461746118042001280B32172E676F6F676C652E70726F746F6275662E5374727563745204646174612289010A105573657253746F72655265717565737412240A0D6F7065726174696F6E54797065180120012809520D6F7065726174696F6E5479706512220A0C6F7267616E697A6174696F6E180220012809520C6F7267616E697A6174696F6E122B0A046461746118032001280B32172E676F6F676C652E70726F746F6275662E537472756374520464617461228A010A115573657253746F7265526573706F6E736512240A0D6F7065726174696F6E54797065180120012809520D6F7065726174696F6E5479706512220A0C6F7267616E697A6174696F6E180220012809520C6F7267616E697A6174696F6E122B0A046461746118032001280B32172E676F6F676C652E70726F746F6275662E53747275637452046461746132440A0F52656D6F74655573657253746F726512310A0B636F6D6D756E6963617465120E2E52656D6F74654D6573736167651A0E2E52656D6F74654D6573736167652801300132480A0C52656D6F746553657276657212380A0F696E766F6B655573657253746F726512112E5573657253746F7265526571756573741A122E5573657253746F7265526573706F6E7365620670726F746F33";

public isolated client class RemoteUserStoreClient {
    *grpc:AbstractClientEndpoint;

    private final grpc:Client grpcClient;

    public isolated function init(string url, *grpc:ClientConfiguration config) returns grpc:Error? {
        self.grpcClient = check new (url, config);
        check self.grpcClient.initStub(self, REMOTE_US_DESC);
    }

    isolated remote function communicate() returns CommunicateStreamingClient|grpc:Error {
        grpc:StreamingClient sClient = check self.grpcClient->executeBidirectionalStreaming("RemoteUserStore/communicate");
        return new CommunicateStreamingClient(sClient);
    }
}

public isolated client class RemoteServerClient {
    *grpc:AbstractClientEndpoint;

    private final grpc:Client grpcClient;

    public isolated function init(string url, *grpc:ClientConfiguration config) returns grpc:Error? {
        self.grpcClient = check new (url, config);
        check self.grpcClient.initStub(self, REMOTE_US_DESC);
    }

    isolated remote function invokeUserStore(UserStoreRequest|ContextUserStoreRequest req) returns UserStoreResponse|grpc:Error {
        map<string|string[]> headers = {};
        UserStoreRequest message;
        if req is ContextUserStoreRequest {
            message = req.content;
            headers = req.headers;
        } else {
            message = req;
        }
        var payload = check self.grpcClient->executeSimpleRPC("RemoteServer/invokeUserStore", message, headers);
        [anydata, map<string|string[]>] [result, _] = payload;
        return <UserStoreResponse>result;
    }

    isolated remote function invokeUserStoreContext(UserStoreRequest|ContextUserStoreRequest req) returns ContextUserStoreResponse|grpc:Error {
        map<string|string[]> headers = {};
        UserStoreRequest message;
        if req is ContextUserStoreRequest {
            message = req.content;
            headers = req.headers;
        } else {
            message = req;
        }
        var payload = check self.grpcClient->executeSimpleRPC("RemoteServer/invokeUserStore", message, headers);
        [anydata, map<string|string[]>] [result, respHeaders] = payload;
        return {content: <UserStoreResponse>result, headers: respHeaders};
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

public client class RemoteUserStoreRemoteMessageCaller {
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

public client class RemoteServerUserStoreResponseCaller {
    private grpc:Caller caller;

    public isolated function init(grpc:Caller caller) {
        self.caller = caller;
    }

    public isolated function getId() returns int {
        return self.caller.getId();
    }

    isolated remote function sendUserStoreResponse(UserStoreResponse response) returns grpc:Error? {
        return self.caller->send(response);
    }

    isolated remote function sendContextUserStoreResponse(ContextUserStoreResponse response) returns grpc:Error? {
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

public type ContextUserStoreRequest record {|
    UserStoreRequest content;
    map<string|string[]> headers;
|};

public type ContextUserStoreResponse record {|
    UserStoreResponse content;
    map<string|string[]> headers;
|};

@protobuf:Descriptor {value: REMOTE_US_DESC}
public type RemoteMessage record {|
    string id = "";
    string operationType = "";
    string organization = "";
    map<anydata> data = {};
|};

@protobuf:Descriptor {value: REMOTE_US_DESC}
public type UserStoreRequest record {|
    string operationType = "";
    string organization = "";
    map<anydata> data = {};
|};

@protobuf:Descriptor {value: REMOTE_US_DESC}
public type UserStoreResponse record {|
    string operationType = "";
    string organization = "";
    map<anydata> data = {};
|};

