import ballerina/log;

// Map to store remote jobs to be processed for each organization.
isolated map<RemoteJob[]> remoteJobs = {};

// Map to store remote responses received from the remote services.
isolated map<RemoteResponse> remoteResponses = {};

// Add a remote job to the organization.
// 
// + organization   - The organization to which the remote job belongs.
// + job            - The remote job to be added.
isolated function addRemoteJob(string organization, RemoteJob job) {

    lock {
        if remoteJobs.hasKey(organization) {
            remoteJobs.get(organization).push(job.clone());
            log:printDebug("Remote job: <" + job.id + "> added to the existing organization: " + organization);
        } else {
            RemoteJob[] orgJobs = [];
            orgJobs.push(job.clone());
            remoteJobs[organization] = orgJobs;
            log:printDebug("Remote job: <" + job.id + "> added to the organization: " + organization);
        }
    }
}

// Get a remote job from the organization.
//
// + organization   - The organization.
// + returns        - The remote job or null if no jobs are available for the organization.
isolated function getRemoteJob(string organization) returns RemoteJob|null {

    RemoteJob|null returnJob = null;

    lock {
        if remoteJobs.hasKey(organization) && remoteJobs.get(organization).length() > 0 {
            RemoteJob job = remoteJobs.get(organization).remove(0);

            log:printDebug("Retrieving remote job for the organization: " + organization 
                + " with id: <" + job.id + ">.");

            returnJob = {
                id: job.id,
                operationType: job.operationType,
                organization: job.organization,
                data: job.data.clone()
            };
        }
    }

    return returnJob;
}

// Add a remote response.
// 
// + response - The remote response to be added.
isolated function addRemoteResponse(RemoteResponse response) {

    lock {
        remoteResponses[response.id] = response.clone();
        log:printDebug("Remote response added with id: <" + response.id + ">.");
    }
}

// Get a remote response.
//
// + id      - The id of the remote response.
// + returns - The remote response or null if no response is available for the id.
isolated function getRemoteResponse(string id) returns RemoteResponse|null {

    RemoteResponse|null remoteResponse = null;

    lock {
        if remoteResponses.hasKey(id) {
            remoteResponse = remoteResponses.remove(id).clone();
        }
    }

    return remoteResponse;
}
