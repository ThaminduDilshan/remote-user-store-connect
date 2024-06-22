import ballerina/http;
import ballerina/log;

// Rest service to receive the user store requests.
public service class ClientRestService {

    *http:Service;

    resource function post invoke\-userstore(RemoteRestRequest req) returns RemoteRESTResponse {
        
        return processUserStoreRESTRequest(req);
    }
}

// Function to process the user store request.
//
// + remoteRequest - The remote REST request.
// + returns - The remote REST response.
function processUserStoreRESTRequest(RemoteRestRequest remoteRequest) returns RemoteRESTResponse {

    string id = remoteRequest.id;
    
    match remoteRequest.operationType {
        DO_AUTHENTICATE => {
            json|error? authData = remoteRequest.data.toJson();

            if authData is error {
                log:printError("Error occurred while parsing the authentication data for the id: " + id);

                UserStoreError usError = {
                    _error: "INVALID_REQUEST",
                    message: "Unable to parse the authentication data."
                };
                
                RemoteRESTResponse remoteResponse = {
                    id: id,
                    operationType: USERSTORE_OPERATION_RESPONSE,
                    data: usError
                };
                return remoteResponse;
            } else {
                json|error? userName = authData.username;
                json|error? password = authData.password;

                if userName is error || password is error {
                    log:printError("Error occurred while parsing username/ password for the id: " + id);
                    
                    UserStoreError usError = {
                        _error: "INVALID_REQUEST",
                        message: "Unable to parse the authentication data."
                    };
                    
                    RemoteRESTResponse remoteResponse = {
                        id: id,
                        operationType: USERSTORE_OPERATION_RESPONSE,
                        data: usError
                    };
                    return remoteResponse;
                }

                string userNameStr = userName.toString();
                string passwordStr = password.toString();

                AuthenticationRequest authRequest = {
                    username: userNameStr,
                    password: passwordStr
                };

                AuthenticationSuccessResponse|AuthenticationFailResponse authResponse = doAuthenticate(authRequest);

                RemoteRESTResponse remoteResponse = {
                    id: id,
                    operationType: USERSTORE_OPERATION_RESPONSE,
                    data: authResponse
                };
                return remoteResponse;
            }
        }
        _ => {
            log:printError("Invalid operation type received for the id: " + id);
            
            UserStoreError usError = {
                _error: "INVALID_REQUEST",
                message: "Invalid operation type received."
            };
            
            RemoteRESTResponse remoteResponse = {
                id: id,
                operationType: USERSTORE_OPERATION_RESPONSE,
                data: usError
            };
            return remoteResponse;
        }
    }
}
