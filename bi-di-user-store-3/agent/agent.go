package main

import (
	"context"
	"io"
	"log"
	"sync"

	pb "bi-di-user-store-3/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	intermediateServerAddress = "localhost:9004"
	organization              = "test_org_1"
	batchSize                 = 5
	noOfMaxConnections        = batchSize * 4
)

func main() {

	log.Println("Starting Local Agent...")

	connectionCount := 0
	mu := sync.Mutex{}

	var wg sync.WaitGroup

	// Start initial no of connections for load handling.
	createBatchConnections(&wg, &connectionCount, &mu)

	wg.Wait()
}

func createBatchConnections(wg *sync.WaitGroup, connectionCount *int, mu *sync.Mutex) {

	for i := 0; i < batchSize; i++ {
		wg.Add(1)
		connectionID := uuid.New().String()
		go startConnection(connectionID, wg, connectionCount, mu)
	}
}

func clearConnection(wg *sync.WaitGroup, connectionCount *int, mu *sync.Mutex) {

	mu.Lock()
	*connectionCount--
	mu.Unlock()
	wg.Done()
}

func startConnection(connectionId string, wg *sync.WaitGroup, connectionCount *int, mu *sync.Mutex) {

	defer clearConnection(wg, connectionCount, mu)

	conn, err := grpc.Dial(intermediateServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewRemoteUserStoreClient(conn)
	stream, err := client.Communicate(context.Background())
	if err != nil {
		log.Fatalf("could not communicate: %v", err)
	}
	defer stream.CloseSend()

	mu.Lock()
	*connectionCount++
	mu.Unlock()
	log.Printf("Agent %s: Connection established. No of active agent connections: %d",
		connectionId, *connectionCount)

	// Send initial connection message with organization ID
	connectionData, _ := structpb.NewStruct(map[string]interface{}{
		"noOfMaxConnections": noOfMaxConnections,
		"noOfMinConnections": batchSize,
	})

	connectMessage := &pb.RemoteMessage{
		OperationType: "CLIENT_CONNECT",
		Id:            uuid.New().String(),
		Organization:  organization,
		Data:          connectionData,
	}
	if err := stream.Send(connectMessage); err != nil {
		log.Fatalf("Agent %s: failed to send connect message: %v", connectionId, err)
	}

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Printf("Agent %s: Stream closed by server", connectionId)
				return
			}
			log.Fatalf("Agent %s: failed to receive: %v", connectionId, err)
		}
		log.Printf("Agent %s: Received request for operation type: %s with id: %s",
			connectionId, req.OperationType, req.Id)

		var response string
		if req.OperationType == "DO_AUTHENTICATE" {
			username := req.Data.Fields["username"].GetStringValue()
			password := req.Data.Fields["password"].GetStringValue()
			response = processAuthenticationRequest(username, password)

			responseData, _ := structpb.NewStruct(map[string]interface{}{
				"response": response,
			})
			if err := stream.Send(&pb.RemoteMessage{Id: req.Id, OperationType: req.OperationType, Organization: req.Organization, Data: responseData}); err != nil {
				log.Fatalf("Agent %s: failed to send response: %v", connectionId, err)
			}
			log.Printf("Agent %s: Sent response for request: %s", connectionId, req.Id)
		} else if req.OperationType == "INCREASE_CONNECTIONS" {
			mu.Lock()
			// Increase the connection count if it is less than the max limit.
			if *connectionCount < noOfMaxConnections {
				log.Printf("Agent %s: Increasing connection count", connectionId)
				mu.Unlock()
				createBatchConnections(wg, connectionCount, mu)
			} else {
				log.Printf("Agent %s: Max connection limit reached", connectionId)
				mu.Unlock()
			}
		} else if req.OperationType == "DECREASE_CONNECTIONS" {
			// TODO
		} else {
			message := req.Data.Fields["message"].GetStringValue()
			response = processUserRequest(message)

			responseData, _ := structpb.NewStruct(map[string]interface{}{
				"response": response,
			})
			if err := stream.Send(&pb.RemoteMessage{Id: req.Id, OperationType: req.OperationType, Organization: req.Organization, Data: responseData}); err != nil {
				log.Fatalf("Agent %s: failed to send response: %v", connectionId, err)
			}
			log.Printf("Agent %s: Sent response for request: %s", connectionId, req.Id)
		}
	}
}

func processUserRequest(request string) string {

	return "Processed: " + request
}

func processAuthenticationRequest(username, password string) string {

	// Implement authentication logic here
	if username == "user1" && password == "user1" {
		return "Authentication successful"
	}
	return "Authentication failed"
}
