type User record {|
    string userId;
    string firstName;
    string lastName;
    string email;
|};

type AuthenticationRequest record {
    string username;
    string password;
};

type AuthenticationSuccessResponse record {
    string username;
    string userId;
    string email;
};

type AuthenticationFailResponse record {
    string username;
    string _error;
    string message;
};

type UserStoreError record {|
    string _error;
    string message;
|};
