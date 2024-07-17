package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	pb "bi-di-user-store-3/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	port              = ":9004"
	maxMsgSize        = 1024 * 1024 * 100 // 100MB
	requestTimeOut    = 15 * time.Second
	agentKeySeperator = "_"
)

const serverId = "server1"

// const agentInstallationToken = "abcdd-1234-efgh-5678"

type agentConnection struct {
	stream pb.UserStoreHubService_CommunicateServer
	inUse  bool
}

type tenantConnection struct {
	tenant           string
	userStore        string
	agentConnections []*agentConnection
	freeConnections  int
	mu               sync.Mutex
}

type server struct {
	pb.UnimplementedUserStoreHubServiceServer
	pb.UnimplementedRemoteUserStoreServiceServer
	mu            sync.Mutex
	agents        map[string]*tenantConnection
	responseChans map[string]chan *pb.UserStoreResponse
	tokenCache    map[string]TokenRecord
}

func (s *server) InvokeUserStore(ctx context.Context, req *pb.UserStoreRequest) (*pb.UserStoreResponse, error) {

	correlationId := req.CorrelationId
	requestID := req.RequestId
	agentKey := req.Tenant + agentKeySeperator + req.UserStore
	startTime := time.Now()

	log.Printf("Received user store request with correlation id: %s, request id: %s", correlationId, requestID)

	if s.agents != nil && s.agents[agentKey] != nil {
		log.Printf("Free connections: %d. Current map: %+v",
			s.agents[agentKey].freeConnections, s.agents[agentKey].agentConnections)
	}

	// Find an available agent connection to handle the request
	var tenantConn *tenantConnection
	var selectedAgentConn *agentConnection

	tenantConn, selectedAgentConn = getAgentConnection(s, agentKey, startTime)

	if selectedAgentConn == nil {
		log.Printf("No available agent connection found for tenant: %s and user store: %s", req.Tenant, req.UserStore)
		return createErrorResponse(req, "No available agent connection"), nil
	}

	s.mu.Lock()

	if s.responseChans == nil {
		s.responseChans = make(map[string]chan *pb.UserStoreResponse)
	}
	responseChan := make(chan *pb.UserStoreResponse, 1)
	s.responseChans[requestID] = responseChan // Store the response channel in the map

	s.mu.Unlock()

	defer func() {
		tenantConn.mu.Lock()
		selectedAgentConn.inUse = false
		tenantConn.freeConnections++
		tenantConn.mu.Unlock()

		s.mu.Lock()
		delete(s.responseChans, requestID) // Clean up the response channel
		s.mu.Unlock()

		close(responseChan)
	}()

	err := selectedAgentConn.stream.Send(&pb.RemoteMessage{
		CorrelationId: correlationId,
		RequestId:     requestID,
		OperationType: req.OperationType,
		Tenant:        req.Tenant,
		UserStore:     req.UserStore,
		Data:          req.Data,
	})

	if err != nil {
		log.Printf("Error sending request to the agent: %v", err)
		return createErrorResponse(req, "Error sending request to the agent"), nil
	}

	// Create a new context with the request timeout
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, requestTimeOut)
	defer timeoutCancel()

	// Wait for the response
	select {
	case resp := <-responseChan: // Wait for the response on the channel
		if resp == nil {
			return createErrorResponse(req, "Error processing request"), nil
		}
		return resp, nil
	case <-timeoutCtx.Done():
		log.Printf("Timeout reached for request: %s", requestID)
		return createErrorResponse(req, "Timeout reached for the request"), nil
	case <-ctx.Done():
		log.Printf("Context done for request: %s", requestID)
		return createErrorResponse(req, "Connection closed with the client"), nil
	}
}

func (s *server) Communicate(stream pb.UserStoreHubService_CommunicateServer) error {

	if s.tokenCache == nil {
		s.tokenCache = make(map[string]TokenRecord)
	}

	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		// return fmt.Errorf("failed to get metadata")
		return status.Errorf(codes.Unauthenticated, "unauthorized. Failed to retrieve metadata")
	}
	tokens := md.Get("authorization")

	if len(tokens) == 0 {
		// return fmt.Errorf("unauthorized. Invalid installation token")
		return status.Errorf(codes.Unauthenticated, "unauthorized. Invalid installation token")
	}

	accessToken := strings.TrimPrefix(tokens[0], "Bearer ")
	accessToken = strings.TrimSpace(accessToken)

	// Validate the installation token.
	savedToken, err := getToken(s, accessToken)
	if err != nil {
		// return fmt.Errorf("unauthorized. Invalid installation token: %v", err)
		return status.Errorf(codes.Unauthenticated, "unauthorized. Invalid installation token: %v", err)
	}

	if savedToken.TOKEN != accessToken {
		// return fmt.Errorf("unauthorized. Invalid installation token")
		return status.Errorf(codes.Unauthenticated, "unauthorized. Invalid installation token")
	} else {
		log.Println("Installation token verified")
	}

	// if len(accessToken) == 0 || accessToken != "Bearer "+agentInstallationToken {
	// 	return fmt.Errorf("Unauthorized. Invalid installation token.")
	// }

	// Receive the initial message to get the tenant ID and user store ID.
	initialMsg, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive initial message: %v", err)
	}

	if initialMsg.OperationType != "CLIENT_CONNECT" {
		return fmt.Errorf("invalid initial operation type: %s", initialMsg.OperationType)
	}

	tenant := initialMsg.Tenant
	userStore := initialMsg.UserStore
	agentKey := tenant + agentKeySeperator + userStore
	agentID := initialMsg.RequestId

	conn := &agentConnection{stream: stream, inUse: false}

	s.mu.Lock()
	if s.agents == nil {
		s.agents = make(map[string]*tenantConnection)
		s.agents[agentKey] = &tenantConnection{
			tenant:           tenant,
			userStore:        userStore,
			agentConnections: []*agentConnection{conn},
			freeConnections:  1,
		}
	} else {
		tenantConn, ok := s.agents[agentKey]
		if ok {
			tenantConn.mu.Lock()
			tenantConn.agentConnections = append(tenantConn.agentConnections, conn)
			tenantConn.freeConnections++
			tenantConn.mu.Unlock()
		} else {
			s.agents[agentKey] = &tenantConnection{
				tenant:           tenant,
				userStore:        userStore,
				agentConnections: []*agentConnection{conn},
				freeConnections:  1,
			}
		}
	}

	if s.responseChans == nil {
		s.responseChans = make(map[string]chan *pb.UserStoreResponse)
	}

	s.mu.Unlock()

	// Send connection success message to the agent.
	connectResponse := &pb.RemoteMessage{
		OperationType: "SERVER_CONNECTED",
		RequestId:     serverId,
		Tenant:        tenant,
		UserStore:     userStore,
		Data:          &structpb.Struct{},
	}

	if err := stream.Send(connectResponse); err != nil {
		return fmt.Errorf("failed to send connection success response to the agent: %s. Error: %v", agentID, err)
	}

	log.Printf("Agent with id: %s connected for tenant: %s and user store: %s", agentID, tenant, userStore)

	for {
		req, err := stream.Recv()
		if err != nil {
			// Remove the connection on error.
			s.removeAgentConnection(agentKey, conn)
			log.Printf("Agent disconnected: %s", agentID)
			return err
		}

		log.Printf("Received message from agent: %s with RequestId: %s, OperationType: %s, Tenant: %s, UserStore: %s", agentID, req.RequestId, req.OperationType, req.Tenant, req.UserStore)

		s.mu.Lock()
		responseChan, exists := s.responseChans[req.RequestId] // Find the corresponding response channel
		s.mu.Unlock()

		if exists {
			responseChan <- &pb.UserStoreResponse{
				CorrelationId: req.CorrelationId,
				RequestId:     req.RequestId,
				OperationType: req.OperationType,
				Data:          req.Data,
			}
		} else {
			log.Printf("No response channel found for request ID: %s", req.RequestId)
		}
	}
}

// Retrieves an agent connection for the given agentKey.
func getAgentConnection(s *server, agentKey string, requestStartTime time.Time) (*tenantConnection, *agentConnection) {

	tenantConn, ok := s.agents[agentKey]

	if !ok {
		log.Printf("No entries found in the agent pool for the agent key: %s", agentKey)
		return nil, nil
	} else if len(tenantConn.agentConnections) == 0 {
		log.Printf("No agent connections found for agent key: %s", agentKey)
		return nil, nil
	}

	return doGetAgentConnection(tenantConn, requestStartTime, 1)
}

func doGetAgentConnection(tenantConn *tenantConnection, requestStartTime time.Time, retrieveAttempt int) (*tenantConnection, *agentConnection) {

	var returnTenantConn *tenantConnection
	var returnAgentConn *agentConnection

	log.Printf("Retrieving agent connection for the tenant: %s in attempt: %d", tenantConn.tenant, retrieveAttempt)

	if time.Since(requestStartTime) > requestTimeOut {
		log.Printf("Timeout reached for tenant: %s in attempt: %d", tenantConn.tenant, retrieveAttempt)
		return nil, nil
	}

	tenantConn.mu.Lock()

	if tenantConn.freeConnections == 0 {
		tenantConn.mu.Unlock()
		log.Printf("No available agent connection found for tenant: %s in attempt: %d",
			tenantConn.tenant, retrieveAttempt)
		return doGetAgentConnection(tenantConn, requestStartTime, retrieveAttempt+1)
	} else {
		for _, conn := range tenantConn.agentConnections {
			if !conn.inUse {
				conn.inUse = true
				tenantConn.freeConnections--
				returnTenantConn = tenantConn
				returnAgentConn = conn
				break
			}
		}

		tenantConn.mu.Unlock()
	}

	return returnTenantConn, returnAgentConn
}

func (s *server) removeAgentConnection(agentKey string, conn *agentConnection) {

	s.mu.Lock()

	tenantConn, ok := s.agents[agentKey]
	if !ok {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	tenantConn.mu.Lock()

	for i, c := range tenantConn.agentConnections {
		if c == conn {
			tenantConn.agentConnections = append(tenantConn.agentConnections[:i],
				tenantConn.agentConnections[i+1:]...)
			tenantConn.freeConnections--
			break
		}
	}

	tenantConn.mu.Unlock()
}

func createErrorResponse(req *pb.UserStoreRequest, message string) *pb.UserStoreResponse {

	errorData, _ := structpb.NewStruct(map[string]interface{}{
		"status":  "FAIL",
		"message": message,
	})
	return &pb.UserStoreResponse{
		CorrelationId: req.CorrelationId,
		RequestId:     req.RequestId,
		OperationType: req.OperationType,
		Data:          errorData,
	}
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
		agents:        make(map[string]*tenantConnection),
		responseChans: make(map[string]chan *pb.UserStoreResponse),
	}

	pb.RegisterRemoteUserStoreServiceServer(grpcServer, s)
	pb.RegisterUserStoreHubServiceServer(grpcServer, s)
	log.Println("Intermediate Server listening on port 9004")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
