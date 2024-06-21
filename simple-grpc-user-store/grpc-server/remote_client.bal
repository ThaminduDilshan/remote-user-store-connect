import ballerina/log;

public function authenticate(AuthenticationRequest authRequest) returns AuthenticationResponse|error {

    RemoteServiceClient ep = check new ("http://localhost:9090");

    RemoteAuthRequest remoteAuthRequest = {
        username: authRequest.username,
        password: authRequest.password
    };

    RemoteAuthResponse|error response = check ep->authenticate(remoteAuthRequest);

    if response is error {
        log:printError("Error occurred while authenticating the user: ", response);

        AuthenticationResponse authResponse = {
            status: "FAIL",
            message: "Error occurred while authenticating the user",
            user: null
        };

        return authResponse;
    }

    if (response.status == "SUCCESS") {
        log:printInfo("User authenticated successfully: ");

        AuthenticatedUser user = {
            userId: response.user.userId,
            username: response.user.username,
            email: response.user.email
        };

        AuthenticationResponse authResponse = {
            status: response.status,
            message: "",
            user: user
        };

        return authResponse;
    } else {
        log:printInfo("User authentication failed: ");

        AuthenticationResponse authResponse = {
            status: response.status,
            message: response.message,
            user: null
        };

        return authResponse;
    }
}
