import ballerina/grpc;
import ballerina/protobuf;

public const string SERVER_DESC = "0A0C7365727665722E70726F746F1A1C676F6F676C652F70726F746F6275662F7374727563742E70726F746F22B2010A0D52656D6F74654D657373616765120E0A02696418012001280952026964121A0A0873747265616D4964180220012809520873747265616D496412220A0C6F7267616E697A6174696F6E180320012809520C6F7267616E697A6174696F6E12240A0D6F7065726174696F6E54797065180420012809520D6F7065726174696F6E54797065122B0A046461746118052001280B32172E676F6F676C652E70726F746F6275662E53747275637452046461746122730A1541757468656E7469636174696F6E52657175657374121A0A08757365726E616D651801200128095208757365726E616D65121A0A0870617373776F7264180220012809520870617373776F726412220A0C6F7267616E697A6174696F6E180320012809520C6F7267616E697A6174696F6E22610A1641757468656E7469636174696F6E526573706F6E7365121A0A08757365726E616D651801200128095208757365726E616D65122B0A046461746118022001280B32172E676F6F676C652E70726F746F6275662E537472756374520464617461323F0A0A4752504353657276657212310A0B636F6D6D756E6963617465120E2E52656D6F74654D6573736167651A0E2E52656D6F74654D6573736167652801300132520A0F5573657253746F7265536572766572123F0A0C61757468656E74696361746512162E41757468656E7469636174696F6E526571756573741A172E41757468656E7469636174696F6E526573706F6E7365620670726F746F33";

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

public isolated client class UserStoreServerClient {
    *grpc:AbstractClientEndpoint;

    private final grpc:Client grpcClient;

    public isolated function init(string url, *grpc:ClientConfiguration config) returns grpc:Error? {
        self.grpcClient = check new (url, config);
        check self.grpcClient.initStub(self, SERVER_DESC);
    }

    isolated remote function authenticate(AuthenticationRequest|ContextAuthenticationRequest req) returns AuthenticationResponse|grpc:Error {
        map<string|string[]> headers = {};
        AuthenticationRequest message;
        if req is ContextAuthenticationRequest {
            message = req.content;
            headers = req.headers;
        } else {
            message = req;
        }
        var payload = check self.grpcClient->executeSimpleRPC("UserStoreServer/authenticate", message, headers);
        [anydata, map<string|string[]>] [result, _] = payload;
        return <AuthenticationResponse>result;
    }

    isolated remote function authenticateContext(AuthenticationRequest|ContextAuthenticationRequest req) returns ContextAuthenticationResponse|grpc:Error {
        map<string|string[]> headers = {};
        AuthenticationRequest message;
        if req is ContextAuthenticationRequest {
            message = req.content;
            headers = req.headers;
        } else {
            message = req;
        }
        var payload = check self.grpcClient->executeSimpleRPC("UserStoreServer/authenticate", message, headers);
        [anydata, map<string|string[]>] [result, respHeaders] = payload;
        return {content: <AuthenticationResponse>result, headers: respHeaders};
    }
}

public isolated client class CommunicateStreamingClient {
    private final grpc:StreamingClient sClient;

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

public isolated client class GRPCServerRemoteMessageCaller {
    private final grpc:Caller caller;

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

public client class UserStoreServerAuthenticationResponseCaller {
    private grpc:Caller caller;

    public isolated function init(grpc:Caller caller) {
        self.caller = caller;
    }

    public isolated function getId() returns int {
        return self.caller.getId();
    }

    isolated remote function sendAuthenticationResponse(AuthenticationResponse response) returns grpc:Error? {
        return self.caller->send(response);
    }

    isolated remote function sendContextAuthenticationResponse(ContextAuthenticationResponse response) returns grpc:Error? {
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

public type ContextAuthenticationRequest record {|
    AuthenticationRequest content;
    map<string|string[]> headers;
|};

public type ContextAuthenticationResponse record {|
    AuthenticationResponse content;
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

@protobuf:Descriptor {value: SERVER_DESC}
public type AuthenticationRequest record {|
    string username = "";
    string password = "";
    string organization = "";
|};

@protobuf:Descriptor {value: SERVER_DESC}
public type AuthenticationResponse record {|
    string username = "";
    map<anydata> data = {};
|};

