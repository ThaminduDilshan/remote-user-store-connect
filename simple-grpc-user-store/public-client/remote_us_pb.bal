import ballerina/grpc;
import ballerina/protobuf;

public const string REMOTE_US_DESC = "0A0F72656D6F74655F75732E70726F746F224B0A1152656D6F74654175746852657175657374121A0A08757365726E616D651802200128095208757365726E616D65121A0A0870617373776F7264180320012809520870617373776F726422560A0A52656D6F74655573657212160A067573657249641801200128095206757365724964121A0A08757365726E616D651802200128095208757365726E616D6512140A05656D61696C1803200128095205656D61696C22670A1252656D6F746541757468526573706F6E736512160A06737461747573180120012809520673746174757312180A076D65737361676518022001280952076D657373616765121F0A047573657218032001280B320B2E52656D6F74655573657252047573657232480A0D52656D6F74655365727669636512370A0C61757468656E74696361746512122E52656D6F746541757468526571756573741A132E52656D6F746541757468526573706F6E7365620670726F746F33";

public isolated client class RemoteServiceClient {
    *grpc:AbstractClientEndpoint;

    private final grpc:Client grpcClient;

    public isolated function init(string url, *grpc:ClientConfiguration config) returns grpc:Error? {
        self.grpcClient = check new (url, config);
        check self.grpcClient.initStub(self, REMOTE_US_DESC);
    }

    isolated remote function authenticate(RemoteAuthRequest|ContextRemoteAuthRequest req) returns RemoteAuthResponse|grpc:Error {
        map<string|string[]> headers = {};
        RemoteAuthRequest message;
        if req is ContextRemoteAuthRequest {
            message = req.content;
            headers = req.headers;
        } else {
            message = req;
        }
        var payload = check self.grpcClient->executeSimpleRPC("RemoteService/authenticate", message, headers);
        [anydata, map<string|string[]>] [result, _] = payload;
        return <RemoteAuthResponse>result;
    }

    isolated remote function authenticateContext(RemoteAuthRequest|ContextRemoteAuthRequest req) returns ContextRemoteAuthResponse|grpc:Error {
        map<string|string[]> headers = {};
        RemoteAuthRequest message;
        if req is ContextRemoteAuthRequest {
            message = req.content;
            headers = req.headers;
        } else {
            message = req;
        }
        var payload = check self.grpcClient->executeSimpleRPC("RemoteService/authenticate", message, headers);
        [anydata, map<string|string[]>] [result, respHeaders] = payload;
        return {content: <RemoteAuthResponse>result, headers: respHeaders};
    }
}

public client class RemoteServiceRemoteAuthResponseCaller {
    private grpc:Caller caller;

    public isolated function init(grpc:Caller caller) {
        self.caller = caller;
    }

    public isolated function getId() returns int {
        return self.caller.getId();
    }

    isolated remote function sendRemoteAuthResponse(RemoteAuthResponse response) returns grpc:Error? {
        return self.caller->send(response);
    }

    isolated remote function sendContextRemoteAuthResponse(ContextRemoteAuthResponse response) returns grpc:Error? {
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

public type ContextRemoteAuthRequest record {|
    RemoteAuthRequest content;
    map<string|string[]> headers;
|};

public type ContextRemoteAuthResponse record {|
    RemoteAuthResponse content;
    map<string|string[]> headers;
|};

@protobuf:Descriptor {value: REMOTE_US_DESC}
public type RemoteUser record {|
    string userId = "";
    string username = "";
    string email = "";
|};

@protobuf:Descriptor {value: REMOTE_US_DESC}
public type RemoteAuthRequest record {|
    string username = "";
    string password = "";
|};

@protobuf:Descriptor {value: REMOTE_US_DESC}
public type RemoteAuthResponse record {|
    string status = "";
    string message = "";
    RemoteUser user = {};
|};

