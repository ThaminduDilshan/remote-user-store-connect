type RemoteClientData record {|
    string clientId;
    string organization;
    map<RemoteClientStreamData> streams;
|};

type RemoteClientStreamData record {|
    string clientId;
    string streamId;
    GRPCServerRemoteMessageCaller caller;
    stream<RemoteMessage, error?> clientStream;
|};
