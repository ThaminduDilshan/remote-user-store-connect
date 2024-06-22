type RemoteJob record {|
    string id;
    string operationType;
    string organization;
    map<anydata> data;
|};

type RemoteResponse record {|
    string id;
    map<anydata> data;
|};

type RemoteClientData record {|
    string id;
    string organization;
|};
