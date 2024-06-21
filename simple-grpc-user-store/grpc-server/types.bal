public type AuthenticationRequest record {|
    string username;
    string password;
|};

type AuthenticatedUser record {|
    string userId;
    string username;
    string email;
|};

public type AuthenticationResponse record {|
    string status;
    string message;
    AuthenticatedUser|null user;
|};
