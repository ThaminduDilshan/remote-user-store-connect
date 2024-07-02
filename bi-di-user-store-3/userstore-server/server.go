package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	pb "bi-di-user-store-3/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

const (
	port       = ":9004"
	maxMsgSize = 1024 * 1024 * 100 // 100MB
)

type agentConnection struct {
	stream       pb.RemoteUserStore_CommunicateServer
	organization string
	inUse        bool
	mu           sync.Mutex
}

type server struct {
	pb.UnimplementedRemoteServerServer
	pb.UnimplementedRemoteUserStoreServer
	mu            sync.Mutex
	agents        map[string][]*agentConnection
	responseChans map[string]chan *pb.UserStoreResponse
}

func (s *server) InvokeUserStore(ctx context.Context, req *pb.UserStoreRequest) (*pb.UserStoreResponse, error) {

	requestID := uuid.New().String() // Generate a UUID for the request ID

	log.Printf("Received InvokeUserStore request: %s", requestID)
	log.Printf("Current agents map: %+v", s.agents)

	// Find an available agent connection to handle the request
	var selectedConn *agentConnection
	s.mu.Lock()
	agentConns, ok := s.agents[req.Organization]
	if ok {
		for _, conn := range agentConns {
			conn.mu.Lock()
			if !conn.inUse {
				conn.inUse = true
				selectedConn = conn
				conn.mu.Unlock()
				break
			}
			conn.mu.Unlock()
		}
	}
	if s.responseChans == nil {
		s.responseChans = make(map[string]chan *pb.UserStoreResponse)
	}
	responseChan := make(chan *pb.UserStoreResponse, 1)
	s.responseChans[requestID] = responseChan // Store the response channel in the map
	s.mu.Unlock()

	if selectedConn == nil {
		return nil, fmt.Errorf("no available agent connection")
	}

	defer func() {
		selectedConn.mu.Lock()
		selectedConn.inUse = false
		selectedConn.mu.Unlock()
		s.mu.Lock()
		delete(s.responseChans, requestID) // Clean up the response channel
		s.mu.Unlock()
		close(responseChan)
	}()

	selectedConn.mu.Lock()
	err := selectedConn.stream.Send(&pb.RemoteMessage{
		Id:            requestID,
		OperationType: req.OperationType,
		Organization:  req.Organization,
		Data:          req.Data,
	})
	selectedConn.mu.Unlock()
	if err != nil {
		log.Printf("Error sending to agent: %v", err)
		return nil, err
	}

	// Wait for the response
	select {
	case resp := <-responseChan: // Wait for the response on the channel
		if resp == nil {
			return nil, fmt.Errorf("error processing request")
		}
		return resp, nil
	case <-ctx.Done():
		log.Printf("Context done for request: %s", requestID)
		return nil, ctx.Err()
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
	conn := &agentConnection{stream: stream, organization: organization, inUse: false}

	s.mu.Lock()
	if s.agents == nil {
		s.agents = make(map[string][]*agentConnection)
	}
	if s.responseChans == nil {
		s.responseChans = make(map[string]chan *pb.UserStoreResponse)
	}
	s.agents[organization] = append(s.agents[organization], conn)
	s.mu.Unlock()

	log.Printf("Agent connected: %s for organization: %s", agentID, organization)

	for {
		req, err := stream.Recv()
		if err != nil {
			s.mu.Lock()
			// Remove the connection on error
			s.removeAgentConnection(organization, conn)
			s.mu.Unlock()
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

func (s *server) removeAgentConnection(organization string, conn *agentConnection) {

	connections := s.agents[organization]
	for i, c := range connections {
		if c == conn {
			s.agents[organization] = append(connections[:i], connections[i+1:]...)
			break
		}
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
		agents:        make(map[string][]*agentConnection),
		responseChans: make(map[string]chan *pb.UserStoreResponse),
	}
	pb.RegisterRemoteServerServer(grpcServer, s)
	pb.RegisterRemoteUserStoreServer(grpcServer, s)
	log.Println("Intermediate Server listening on port 9004")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
