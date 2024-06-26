import ballerina/log;

// Define the server endpoint
final string serverEndpoint = "http://localhost:9092";

public function main() returns error? {
    // Create a sample authentication request with hardcoded data
    map<anydata> authData = {};
    authData["username"] = {string_value: "user1"};
    authData["password"] = {string_value: "user1"};

    UserStoreRequest request = {
        operationType: "DO_AUTHENTICATE",
        organization: "test_org_1",
        data: authData
    };

    // Create a gRPC client for the RemoteUserStore service
    RemoteServerClient ep = check new(serverEndpoint);

    // Send the request to the server and receive the response
    var response = ep->invokeUserStore(request);
    if (response is UserStoreResponse) {
        log:printInfo("Response received from server: " + response.toString());
    } else {
        log:printError("Failed to receive response from server: ", response);
    }
}