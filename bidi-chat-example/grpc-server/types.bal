
type RemoteJob record {|
    string id;
    string organization;
    string name;
    string message;
|};

type RemoteResult record {|
    string id;
    string organization;
    string message;
    string status;
|};
