import ballerina/grpc;
import ballerina/protobuf;

public const string REMOTE_US_DESC = "0A0F72656D6F74655F75732E70726F746F1A1C676F6F676C652F70726F746F6275662F7374727563742E70726F746F22640A0D52656D6F74655265717565737412120A047575696418012001280952047575696412120A0474797065180220012809520474797065122B0A046461746118032001280B32172E676F6F676C652E70726F746F6275662E5374727563745204646174612283010A0E52656D6F7465526573706F6E736512120A047575696418012001280952047575696412160A06737461747573180220012809520673746174757312180A076D65737361676518032001280952076D657373616765122B0A046461746118042001280B32172E676F6F676C652E70726F746F6275662E53747275637452046461746122730A1541757468656E7469636174696F6E52657175657374121A0A08757365726E616D651801200128095208757365726E616D65121A0A0870617373776F7264180220012809520870617373776F726412220A0C6F7267616E697A6174696F6E180320012809520C6F7267616E697A6174696F6E22770A1641757468656E7469636174696F6E526573706F6E736512160A06737461747573180120012809520673746174757312180A076D65737361676518022001280952076D657373616765122B0A046461746118032001280B32172E676F6F676C652E70726F746F6275662E5374727563745204646174613288010A1152656D6F7465426944695365727669636512320A0B636F6D6D756E6963617465120F2E52656D6F7465526573706F6E73651A0E2E52656D6F74655265717565737428013001123F0A0C61757468656E74696361746512162E41757468656E7469636174696F6E526571756573741A172E41757468656E7469636174696F6E526573706F6E7365620670726F746F33";

public isolated client class RemoteBiDiServiceClient {
    *grpc:AbstractClientEndpoint;

    private final grpc:Client grpcClient;

    public isolated function init(string url, *grpc:ClientConfiguration config) returns grpc:Error? {
        self.grpcClient = check new (url, config);
        check self.grpcClient.initStub(self, REMOTE_US_DESC);
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
        var payload = check self.grpcClient->executeSimpleRPC("RemoteBiDiService/authenticate", message, headers);
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
        var payload = check self.grpcClient->executeSimpleRPC("RemoteBiDiService/authenticate", message, headers);
        [anydata, map<string|string[]>] [result, respHeaders] = payload;
        return {content: <AuthenticationResponse>result, headers: respHeaders};
    }

    isolated remote function communicate() returns CommunicateStreamingClient|grpc:Error {
        grpc:StreamingClient sClient = check self.grpcClient->executeBidirectionalStreaming("RemoteBiDiService/communicate");
        return new CommunicateStreamingClient(sClient);
    }
}

public client class CommunicateStreamingClient {
    private grpc:StreamingClient sClient;

    isolated function init(grpc:StreamingClient sClient) {
        self.sClient = sClient;
    }

    isolated remote function sendRemoteResponse(RemoteResponse message) returns grpc:Error? {
        return self.sClient->send(message);
    }

    isolated remote function sendContextRemoteResponse(ContextRemoteResponse message) returns grpc:Error? {
        return self.sClient->send(message);
    }

    isolated remote function receiveRemoteRequest() returns RemoteRequest|grpc:Error? {
        var response = check self.sClient->receive();
        if response is () {
            return response;
        } else {
            [anydata, map<string|string[]>] [payload, _] = response;
            return <RemoteRequest>payload;
        }
    }

    isolated remote function receiveContextRemoteRequest() returns ContextRemoteRequest|grpc:Error? {
        var response = check self.sClient->receive();
        if response is () {
            return response;
        } else {
            [anydata, map<string|string[]>] [payload, headers] = response;
            return {content: <RemoteRequest>payload, headers: headers};
        }
    }

    isolated remote function sendError(grpc:Error response) returns grpc:Error? {
        return self.sClient->sendError(response);
    }

    isolated remote function complete() returns grpc:Error? {
        return self.sClient->complete();
    }
}

public client class RemoteBiDiServiceAuthenticationResponseCaller {
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

public client class RemoteBiDiServiceRemoteRequestCaller {
    private grpc:Caller caller;

    public isolated function init(grpc:Caller caller) {
        self.caller = caller;
    }

    public isolated function getId() returns int {
        return self.caller.getId();
    }

    isolated remote function sendRemoteRequest(RemoteRequest response) returns grpc:Error? {
        return self.caller->send(response);
    }

    isolated remote function sendContextRemoteRequest(ContextRemoteRequest response) returns grpc:Error? {
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

public type ContextRemoteResponseStream record {|
    stream<RemoteResponse, error?> content;
    map<string|string[]> headers;
|};

public type ContextRemoteRequestStream record {|
    stream<RemoteRequest, error?> content;
    map<string|string[]> headers;
|};

public type ContextRemoteResponse record {|
    RemoteResponse content;
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

public type ContextRemoteRequest record {|
    RemoteRequest content;
    map<string|string[]> headers;
|};

@protobuf:Descriptor {value: REMOTE_US_DESC}
public type RemoteResponse record {|
    string uuid = "";
    string status = "";
    string message = "";
    map<anydata> data = {};
|};

@protobuf:Descriptor {value: REMOTE_US_DESC}
public type AuthenticationRequest record {|
    string username = "";
    string password = "";
    string organization = "";
|};

@protobuf:Descriptor {value: REMOTE_US_DESC}
public type AuthenticationResponse record {|
    string status = "";
    string message = "";
    map<anydata> data = {};
|};

@protobuf:Descriptor {value: REMOTE_US_DESC}
public type RemoteRequest record {|
    string uuid = "";
    string 'type = "";
    map<anydata> data = {};
|};

