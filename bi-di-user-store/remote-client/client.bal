import ballerina/log;
import ballerina/http;
import ballerina/lang.runtime;

// Server variables.
final string serverEndpoint = "http://localhost:9090";

// Client variables.
final int clientRESTPort = 9095;
final string client_id = "client_001";
final string organization = "test_org_1";

configurable string clientMode = "PRIVATE";

public function main() returns error? {

    log:printInfo("Starting the remote user store client...");

    if clientMode == PRIVATE {
        // Starting the client in private mode. I.e. a GRPC client.
        log:printInfo("Client mode is set to: " + PRIVATE);

        RemoteUserStoreClient ep = check new (serverEndpoint);
        CommunicateStreamingClient streamingClient = check ep->communicate();

        // Start receiving remote requests in another strand.
        future<error?> f1 = start receiveRemoteRequests(streamingClient);

        // Send a client connection request to the server.
        RemoteMessage connectionRequest = {
            operationType: CLIENT_CONNECT,
            id: client_id,
            organization: organization
        };
        check streamingClient->sendRemoteMessage(connectionRequest);

        // Wait for the receiveRemoteRequests function to complete.
        check wait f1;
    } else if clientMode == PUBLIC {
        // Starting the client in public mode. I.e. a REST client.
        log:printInfo("Client mode is set to: " + PUBLIC);

        http:Listener htpListener = check new(clientRESTPort);
        http:Service clientRESTService = new ClientRestService();

        check htpListener.attach(clientRESTService);
        check htpListener.'start();
        
        runtime:registerListener(htpListener);
    } else {
        log:printError("Invalid client mode: " + clientMode);
    }
}

// Function to authenticate the user.
//
// + authRequest - The authentication request.
// + returns     - The authentication response.
function doAuthenticate(AuthenticationRequest authRequest) 
        returns AuthenticationSuccessResponse|AuthenticationFailResponse {

    if USER_CREDENTIALS.hasKey(authRequest.username) 
            && USER_CREDENTIALS.get(authRequest.username) == authRequest.password {
        
        User user = USER_DATA.get(authRequest.username);

        AuthenticationSuccessResponse response = {
            username: authRequest.username,
            userId: user.userId,
            email: user.email
        };

        return response;
    } else {
        AuthenticationFailResponse response = {
            username: authRequest.username,
            _error: "INVALID CREDENTIALS",
            message: "Provided username or password is incorrect. Please try again."
        };

        return response;
    }
}
