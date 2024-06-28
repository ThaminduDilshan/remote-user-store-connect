import ballerina/log;

// Define the server endpoint
final string serverEndpoint = "http://localhost:9092";

public function main() returns error? {

    // Create a sample authentication request with hardcoded data.
    AuthenticationRequest request = {
        username: "user1",
        password: "user1",
        organization: "test_org_1"
    };

    // Create a gRPC client for the UserStoreServer service.
    UserStoreServerClient ep = check new(serverEndpoint);

    // Send the request to the server and receive the response
    var response = ep->authenticate(request);

    if (response is AuthenticationResponse) {
        log:printInfo("Response received from server: " + response.toString());
    } else {
        log:printError("Failed to receive response from server: ", response);
    }
}
