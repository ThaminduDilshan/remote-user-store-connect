import ballerina/grpc;
import ballerina/log;

@grpc:Descriptor {
    value: REMOTE_US_DESC
}
service "RemoteService" on new grpc:Listener(9090) {
    
    remote function authenticate(RemoteAuthRequest authRequest) returns RemoteAuthResponse {

        log:printInfo("Received an authenticate request for the user: " + authRequest.username);
        
        if (authRequest.username == "test1" && authRequest.password == "test1") {
            return {
                status: "SUCCESS",
                message: "",
                user: {
                    userId: "1",
                    username: "test1",
                    email: "test1@wso2.com"
                }
            };
        } else {
            return {
                status: "FAILED",
                message: "Invalid credentials",
                user: {}
            };
        }
    }
}
