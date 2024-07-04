package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	pb "bi-di-user-store-3/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	port           = ":9004"
	maxMsgSize     = 1024 * 1024 * 100 // 100MB
	requestTimeOut = 15 * time.Second
)

type connectionMetaData struct {
	requestedConnections int
	previousConnCount    int
	batchSize            int
}

type agentConnection struct {
	stream pb.RemoteUserStore_CommunicateServer
	inUse  bool
}

type organizationConnection struct {
	organization            string
	agentConnections        []*agentConnection
	freeConnections         int
	maxSupportedConnections int
	connectionMetaData      connectionMetaData
	mu                      sync.Mutex
}

type server struct {
	pb.UnimplementedRemoteServerServer
	pb.UnimplementedRemoteUserStoreServer
	mu            sync.Mutex
	agents        map[string]*organizationConnection
	responseChans map[string]chan *pb.UserStoreResponse
}

func (s *server) InvokeUserStore(ctx context.Context, req *pb.UserStoreRequest) (*pb.UserStoreResponse, error) {

	requestID := uuid.New().String() // Generate a UUID for the request ID
	startTime := time.Now()

	log.Printf("Received user store request. Assigned the id: %s", requestID)

	if s.agents != nil && s.agents[req.Organization] != nil {
		log.Printf("Free connections: %d. Current map: %+v",
			s.agents[req.Organization].freeConnections, s.agents[req.Organization].agentConnections)
	}

	// Find an available agent connection to handle the request
	var organizationConn *organizationConnection
	var selectedAgentConn *agentConnection

	organizationConn, selectedAgentConn = getAgentConnection(s, req.Organization, startTime)

	if selectedAgentConn == nil {

		log.Printf("No available agent connection found for organization: %s", req.Organization)

		errorData, _ := structpb.NewStruct(map[string]interface{}{
			"status":  "FAIL",
			"message": "No available agent connection",
		})

		return &pb.UserStoreResponse{
			OperationType: req.OperationType,
			Organization:  req.Organization,
			Data:          errorData,
		}, nil
	}

	s.mu.Lock()

	if s.responseChans == nil {
		s.responseChans = make(map[string]chan *pb.UserStoreResponse)
	}
	responseChan := make(chan *pb.UserStoreResponse, 1)
	s.responseChans[requestID] = responseChan // Store the response channel in the map

	s.mu.Unlock()

	defer func() {
		organizationConn.mu.Lock()
		selectedAgentConn.inUse = false
		organizationConn.freeConnections++
		organizationConn.mu.Unlock()

		s.mu.Lock()
		delete(s.responseChans, requestID) // Clean up the response channel
		s.mu.Unlock()

		close(responseChan)
	}()

	err := selectedAgentConn.stream.Send(&pb.RemoteMessage{
		Id:            requestID,
		OperationType: req.OperationType,
		Organization:  req.Organization,
		Data:          req.Data,
	})

	if err != nil {
		log.Printf("Error sending request to the agent: %v", err)

		errorData, _ := structpb.NewStruct(map[string]interface{}{
			"status":  "FAIL",
			"message": "Error sending request to the agent",
		})

		return &pb.UserStoreResponse{
			OperationType: req.OperationType,
			Organization:  req.Organization,
			Data:          errorData,
		}, nil
	}

	// Create a new context with the request timeout
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, requestTimeOut)
	defer timeoutCancel()

	// Wait for the response
	select {
	case resp := <-responseChan: // Wait for the response on the channel
		if resp == nil {
			errorData, _ := structpb.NewStruct(map[string]interface{}{
				"status":  "FAIL",
				"message": "Error processing request",
			})

			return &pb.UserStoreResponse{
				OperationType: req.OperationType,
				Organization:  req.Organization,
				Data:          errorData,
			}, nil
		}

		return resp, nil
	case <-timeoutCtx.Done():
		log.Printf("Timeout reached for request: %s", requestID)

		errorData, _ := structpb.NewStruct(map[string]interface{}{
			"status":  "FAIL",
			"message": "Timeout reached for the request",
		})

		return &pb.UserStoreResponse{
			OperationType: req.OperationType,
			Organization:  req.Organization,
			Data:          errorData,
		}, nil
	case <-ctx.Done():
		log.Printf("Context done for request: %s", requestID)

		errorData, _ := structpb.NewStruct(map[string]interface{}{
			"status":  "FAIL",
			"message": "Connection closed with the client",
		})

		return &pb.UserStoreResponse{
			OperationType: req.OperationType,
			Organization:  req.Organization,
			Data:          errorData,
		}, nil
	}
}

func (s *server) Communicate(stream pb.RemoteUserStore_CommunicateServer) error {

	// Receive the initial message to get the organization ID
	initialMsg, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive initial message: %v", err)
	}

	if initialMsg.OperationType != "CLIENT_CONNECT" {
		return fmt.Errorf("invalid initial operation type: %s", initialMsg.OperationType)
	}

	organization := initialMsg.Organization
	agentID := initialMsg.Id

	// Extract the noOfMaxConnections from the initial message data
	data := initialMsg.Data.AsMap()
	noOfMaxConnections, okMax := data["noOfMaxConnections"].(float64)
	noOfMinConnections, okMin := data["noOfMinConnections"].(float64)

	if !okMax {
		log.Printf("noOfMaxConnections not found or not a number in initial message data")
	}
	if !okMin {
		log.Printf("noOfMinConnections not found or not a number in initial message data")
	}

	conn := &agentConnection{stream: stream, inUse: false}

	s.mu.Lock()
	if s.agents == nil {
		s.agents = make(map[string]*organizationConnection)

		s.agents[organization] = &organizationConnection{
			organization:            organization,
			agentConnections:        []*agentConnection{conn},
			freeConnections:         1,
			maxSupportedConnections: int(noOfMaxConnections),
			connectionMetaData: connectionMetaData{
				batchSize: int(noOfMinConnections),
			},
		}
	} else {
		organizationConn, ok := s.agents[organization]

		if ok {
			organizationConn.mu.Lock()
			organizationConn.agentConnections = append(organizationConn.agentConnections, conn)
			organizationConn.freeConnections++
			organizationConn.mu.Unlock()
		} else {
			s.agents[organization] = &organizationConnection{
				organization:            organization,
				agentConnections:        []*agentConnection{conn},
				freeConnections:         1,
				maxSupportedConnections: int(noOfMaxConnections),
				connectionMetaData: connectionMetaData{
					batchSize: int(noOfMinConnections),
				},
			}
		}
	}

	if s.responseChans == nil {
		s.responseChans = make(map[string]chan *pb.UserStoreResponse)
	}

	s.mu.Unlock()

	log.Printf("Agent connected: %s for organization: %s", agentID, organization)

	for {
		req, err := stream.Recv()

		if err != nil {
			// Remove the connection on error.
			s.removeAgentConnection(organization, conn)
			log.Printf("Agent disconnected: %s", agentID)
			return err
		}

		log.Printf("Received message from agent: %s, OperationType: %s, Organization: %s, Data: %v", req.Id, req.OperationType, req.Organization, req.Data)

		s.mu.Lock()
		responseChan, exists := s.responseChans[req.Id] // Find the corresponding response channel
		s.mu.Unlock()

		if exists {
			responseChan <- &pb.UserStoreResponse{
				OperationType: req.OperationType,
				Organization:  req.Organization,
				Data:          req.Data,
			}
		} else {
			log.Printf("No response channel found for request ID: %s", req.Id)
		}
	}
}

func getAgentConnection(s *server, organization string, requestStartTime time.Time) (*organizationConnection, *agentConnection) {

	organizationConn, ok := s.agents[organization]

	if !ok {
		log.Printf("Organization not found in the agent pool: %s", organization)
		return nil, nil
	} else if len(organizationConn.agentConnections) == 0 {
		log.Printf("No agent connections found for organization: %s", organization)
		return nil, nil
	}

	return doGetAgentConnection(organizationConn, requestStartTime, 1)
}

func doGetAgentConnection(organizationConn *organizationConnection, requestStartTime time.Time, retrieveAttempt int) (*organizationConnection, *agentConnection) {

	var returnOrgConn *organizationConnection
	var returnAgentConn *agentConnection

	log.Printf("Retrieving agent connection for organization: %s in attempt: %d", organizationConn.organization, retrieveAttempt)

	if time.Since(requestStartTime) > requestTimeOut {
		log.Printf("Timeout reached for organization: %s in attempt: %d", organizationConn.organization, retrieveAttempt)
		return nil, nil
	}

	organizationConn.mu.Lock()

	if organizationConn.freeConnections == 0 {
		organizationConn.mu.Unlock()
		log.Printf("No available agent connection found for organization: %s in attempt: %d",
			organizationConn.organization, retrieveAttempt)
		return doGetAgentConnection(organizationConn, requestStartTime, retrieveAttempt+1)
	} else {
		for _, conn := range organizationConn.agentConnections {
			if !conn.inUse {
				conn.inUse = true
				organizationConn.freeConnections--
				returnOrgConn = organizationConn
				returnAgentConn = conn
				break
			}
		}

		// If the picked one is the last left connection, try to to increase the connections.
		if organizationConn.freeConnections == 0 {
			sendConnectionIncrementRequest(organizationConn, returnAgentConn)
		}
	}

	organizationConn.mu.Unlock()

	return returnOrgConn, returnAgentConn
}

func sendConnectionIncrementRequest(orgConn *organizationConnection, agentConn *agentConnection) {

	// We assume at this point, org lock is already in possession.

	// Prevent sending increment request when one request is already being processed.
	totalExpectedConnCount := orgConn.connectionMetaData.previousConnCount + orgConn.connectionMetaData.requestedConnections

	if totalExpectedConnCount != 0 && len(orgConn.agentConnections) == totalExpectedConnCount {
		orgConn.connectionMetaData.previousConnCount = totalExpectedConnCount
		orgConn.connectionMetaData.requestedConnections = 0
	}

	if len(orgConn.agentConnections) < orgConn.maxSupportedConnections && orgConn.connectionMetaData.requestedConnections == 0 {

		orgConn.connectionMetaData.previousConnCount = len(orgConn.agentConnections)
		orgConn.connectionMetaData.requestedConnections = orgConn.connectionMetaData.batchSize

		err := agentConn.stream.Send(&pb.RemoteMessage{
			Id:            "",
			OperationType: "INCREASE_CONNECTIONS",
			Organization:  orgConn.organization,
			Data:          nil,
		})

		if err != nil {
			log.Printf("Error sending connection increment request to the agent: %v", err)
		}
	}
}

// // TODO: Simple demo approach.
// func sendConnectionDecrementRequest(orgConn *organizationConnection, agentConn *agentConnection) {

// 	if len(orgConn.agentConnections) > 2*orgConn.minSupportedConnections {
// 		err := agentConn.stream.Send(&pb.RemoteMessage{
// 			Id:            "",
// 			OperationType: "DECREASE_CONNECTIONS",
// 			Organization:  orgConn.organization,
// 			Data:          nil,
// 		})

// 		if err != nil {
// 			log.Printf("Error sending connection decrement request to the agent: %v", err)
// 		}
// 	}
// }

func (s *server) removeAgentConnection(organization string, conn *agentConnection) {

	s.mu.Lock()

	organizationConn, ok := s.agents[organization]
	if !ok {
		s.mu.Unlock()
		return
	}

	organizationConn.mu.Lock()

	for i, c := range organizationConn.agentConnections {
		if c == conn {
			organizationConn.agentConnections = append(organizationConn.agentConnections[:i],
				organizationConn.agentConnections[i+1:]...)
			organizationConn.freeConnections--
			break
		}
	}

	organizationConn.mu.Unlock()
	s.mu.Unlock()
}

func main() {

	log.Println("Starting Intermediate Server...")

	lis, err := net.Listen("tcp", port)

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
	)

	s := &server{
		agents:        make(map[string]*organizationConnection),
		responseChans: make(map[string]chan *pb.UserStoreResponse),
	}

	pb.RegisterRemoteServerServer(grpcServer, s)
	pb.RegisterRemoteUserStoreServer(grpcServer, s)
	log.Println("Intermediate Server listening on port 9004")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
