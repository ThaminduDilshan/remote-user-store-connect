import ballerina/http;

service /call\-remote on new http:Listener(8090) {

    resource function post authenticate(AuthenticationRequest req) returns AuthenticationResponse {
        
        AuthenticationResponse|error authResponse = authenticate(req);

        if authResponse is error {
            return {
                status: "FAIL",
                message: "Something went wrong while authenticating the user",
                user: null
            };
        }

        return authResponse;
    }
}
