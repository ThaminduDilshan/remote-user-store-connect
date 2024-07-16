package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	pb "bi-di-user-store-3/proto"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	port           = ":9004"
	maxMsgSize     = 1024 * 1024 * 100 // 100MB
	requestTimeOut = 15 * time.Second
)

type agentConnection struct {
	stream pb.UserStoreHubService_CommunicateServer
	inUse  bool
	wg     sync.WaitGroup
}

type organizationConnection struct {
	organization     string
	agentConnections []*agentConnection
	freeConnections  int
	mu               sync.Mutex
}

type server struct {
	pb.UnimplementedUserStoreHubServiceServer
	pb.UnimplementedRemoteUserStoreServiceServer
	mu     sync.Mutex
	agents map[string]*organizationConnection
}

func (s *server) InvokeUserStore(ctx context.Context, req *pb.UserStoreRequest) (*pb.UserStoreResponse, error) {

	requestID := req.Id
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
		return createErrorResponse(req, "No available agent connection"), nil
	}

	defer func() {
		organizationConn.mu.Lock()
		selectedAgentConn.inUse = false
		organizationConn.freeConnections++
		organizationConn.mu.Unlock()
	}()

	err := selectedAgentConn.stream.Send(&pb.RemoteMessage{
		Id:            requestID,
		OperationType: req.OperationType,
		Organization:  req.Organization,
		Data:          req.Data,
	})

	if err != nil {
		log.Printf("Error sending request to the agent: %v", err)
		return createErrorResponse(req, "Error sending request to the agent"), nil
	}

	// TODO: We should be able to listen for the response here without response channels.

	for {
		response, err := selectedAgentConn.stream.Recv()
		if err != nil {
			log.Printf("Error receiving response from the agent: %v", err)
			return createErrorResponse(req, "Error receiving response from the agent"), nil
		}

		log.Printf("Received message from agent: %s, OperationType: %s, Organization: %s, Data: %v", req.Id, req.OperationType, req.Organization, req.Data)

		return &pb.UserStoreResponse{
			Id:            response.Id,
			OperationType: response.OperationType,
			Organization:  response.Organization,
			Data:          response.Data,
		}, nil
	}
}

func (s *server) Communicate(stream pb.UserStoreHubService_CommunicateServer) error {

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
	conn := &agentConnection{stream: stream, inUse: false}

	s.mu.Lock()
	if s.agents == nil {
		s.agents = make(map[string]*organizationConnection)
		s.agents[organization] = &organizationConnection{
			organization:     organization,
			agentConnections: []*agentConnection{conn},
			freeConnections:  1,
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
				organization:     organization,
				agentConnections: []*agentConnection{conn},
				freeConnections:  1,
			}
		}
	}

	s.mu.Unlock()

	log.Printf("Agent connected: %s for organization: %s", agentID, organization)

	var wg sync.WaitGroup
	wg.Add(1)

	wg.Wait()
	return nil

	// for {
	// 	log.Printf("Alive...: %s", agentID)
	// }

	// for {
	// 	req, err := stream.Recv()
	// 	if err != nil {
	// 		// Remove the connection on error.
	// 		s.removeAgentConnection(organization, conn)
	// 		log.Printf("Agent disconnected: %s", agentID)
	// 		return err
	// 	}

	// 	log.Printf("Received message from agent: %s, OperationType: %s, Organization: %s, Data: %v", req.Id, req.OperationType, req.Organization, req.Data)
	// }
}

// Retrieves an agent connection for the given organization
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

		organizationConn.mu.Unlock()
	}

	return returnOrgConn, returnAgentConn
}

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
}

func createErrorResponse(req *pb.UserStoreRequest, message string) *pb.UserStoreResponse {

	errorData, _ := structpb.NewStruct(map[string]interface{}{
		"status":  "FAIL",
		"message": message,
	})
	return &pb.UserStoreResponse{
		Id:            req.Id,
		OperationType: req.OperationType,
		Organization:  req.Organization,
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
		agents: make(map[string]*organizationConnection),
	}

	pb.RegisterRemoteUserStoreServiceServer(grpcServer, s)
	pb.RegisterUserStoreHubServiceServer(grpcServer, s)
	log.Println("Intermediate Server listening on port 9004")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
