package main

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	pb "bi-di-user-store-3/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	intermediateServerAddress = "localhost:9004"
	organization              = "test_org_1"
	idleConnections           = 10
	batchConnectionSize       = 5
	noOfMaxConnections        = batchConnectionSize * 6
	statusCheckInterval       = 30 * time.Second
)

func main() {

	log.Println("Starting Local Agent...")

	var wg sync.WaitGroup
	connectionCount := 0
	mu := sync.Mutex{}

	// Start initial no of connections for load handling.
	createBatchConnections(idleConnections, &wg, &connectionCount, &mu)

	wg.Wait()
}

func clearConnection(wg *sync.WaitGroup, connectionCount *int, mu *sync.Mutex) {

	mu.Lock()
	*connectionCount--
	mu.Unlock()
	wg.Done()
}

func createBatchConnections(batchSize int, wg *sync.WaitGroup, connectionCount *int, mu *sync.Mutex) {

	log.Printf("Creating a batch of %d connections...", batchSize)

	for i := 0; i < batchSize; i++ {
		wg.Add(1)
		connectionID := uuid.New().String()
		go startConnection(connectionID, wg, connectionCount, mu)
	}
}

func startConnection(connectionId string, wg *sync.WaitGroup, connectionCount *int, mu *sync.Mutex) {

	defer clearConnection(wg, connectionCount, mu)

	terminateFlag := false

	conn, err := grpc.NewClient(intermediateServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logFatal(connectionId, "", "Failed to connect to the server:", err)
	}
	defer conn.Close()

	client := pb.NewUserStoreHubServiceClient(conn)
	stream, err := client.Communicate(context.Background())
	if err != nil {
		logFatal(connectionId, "", "Failed to create a stream for the communication:", err)
	}
	defer stream.CloseSend()

	mu.Lock()
	*connectionCount++
	mu.Unlock()

	logInfo(connectionId, "", "A new connection established. No of active connections:", *connectionCount)

	// Send initial connection message with organization ID
	connectionData, _ := structpb.NewStruct(map[string]interface{}{
		"idleConnections":     idleConnections,
		"maxConnections":      noOfMaxConnections,
		"batchConnectionSize": batchConnectionSize,
	})

	// Send initial connection message with organization ID
	connectMessage := &pb.RemoteMessage{
		OperationType: "CLIENT_CONNECT",
		Id:            uuid.New().String(),
		Organization:  organization,
		Data:          connectionData,
	}
	if err := stream.Send(connectMessage); err != nil {
		logFatal(connectionId, "", "Failed to send connect message to the server:", err)
	}

	go sendPeriodicStatusCheck(stream, connectionId, &terminateFlag)

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				logInfo(connectionId, "", "Stream closed by the server.")
				return
			}
			logFatal(connectionId, "", "Failed to receive message from the server:", err)
		}

		logInfo(connectionId, req.Id, "Received request for operation type:", req.OperationType)

		var response string

		if req.OperationType == "INCREASE_CONNECTIONS" {
			mu.Lock()
			// Increase the connection count if it is less than the max limit.
			if *connectionCount < noOfMaxConnections {
				logInfo(connectionId, req.Id, "Increasing connection count")
				mu.Unlock()
				createBatchConnections(batchConnectionSize, wg, connectionCount, mu)
			} else {
				logInfo(connectionId, req.Id, "Server has reached the max connection limit.")
				mu.Unlock()
			}
		} else if req.OperationType == "DECREASE_CONNECTIONS" {
			mu.Lock()
			if *connectionCount > idleConnections {
				logInfo(connectionId, req.Id, "Terminating the current connection to decrease the connection count.")
				sendDisconnect(stream, connectionId)
				terminateFlag = true
				mu.Unlock()
				return
			}
			mu.Unlock()
		} else if req.OperationType == "DO_AUTHENTICATE" {
			username := req.Data.Fields["username"].GetStringValue()
			password := req.Data.Fields["password"].GetStringValue()
			response = processAuthenticationRequest(username, password)

			responseData, _ := structpb.NewStruct(map[string]interface{}{
				"response": response,
			})
			if err := stream.Send(&pb.RemoteMessage{Id: req.Id, OperationType: "US_RESPONSE", Organization: req.Organization, Data: responseData}); err != nil {
				logFatal(connectionId, req.Id, "Failed to send response to the server:", err)
			}
			logInfo(connectionId, req.Id, "Sent response for the request:", req.Id)
		} else {
			message := req.Data.Fields["message"].GetStringValue()
			response = processUserRequest(message)

			responseData, _ := structpb.NewStruct(map[string]interface{}{
				"response": response,
			})
			if err := stream.Send(&pb.RemoteMessage{Id: req.Id, OperationType: "US_RESPONSE", Organization: req.Organization, Data: responseData}); err != nil {
				logFatal(connectionId, req.Id, "Failed to send response to the server:", err)
			}
			logInfo(connectionId, req.Id, "Sent response for the request:", req.Id)
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

func sendPeriodicStatusCheck(stream pb.UserStoreHubService_CommunicateClient, connectionId string, terminateFlag *bool) {

	logInfo(connectionId, "", "Starting periodic status check service...")

	lastStatusCheckTime := time.Now()

	statusMessage := &pb.RemoteMessage{
		OperationType: "STATUS_CHECK",
		Id:            connectionId,
		Organization:  organization,
		Data:          nil,
	}

	for {
		if *terminateFlag {
			logInfo(connectionId, "", "Terminating the status check service.")
			return
		} else if time.Since(lastStatusCheckTime) >= statusCheckInterval {
			logInfo(connectionId, "", "Sending status check message to the server...")
			if err := stream.Send(statusMessage); err != nil {
				logError(connectionId, "", "Failed to send status message to the server:", err)
			}
			lastStatusCheckTime = time.Now()
		} else {
			time.Sleep(2 * time.Second)
		}
	}
}

func sendDisconnect(stream pb.UserStoreHubService_CommunicateClient, connectionId string) {

	logInfo(connectionId, "", "Sending disconnect message to the server...")

	discMessage := &pb.RemoteMessage{
		OperationType: "DISCONNECTING",
		Id:            connectionId,
		Organization:  organization,
		Data:          nil,
	}

	if err := stream.Send(discMessage); err != nil {
		logError(connectionId, "", "Failed to send disconnect message to the server:", err)
	}
}

func logInfo(connectionId string, correlationId string, message string, data ...any) {

	if correlationId == "" {
		log.Printf("[conn: %s] %s %v", connectionId, message, data)
	} else {
		log.Printf("[conn: %s][%s] %s %v", connectionId, correlationId, message, data)
	}
}

func logError(connectionId string, correlationId string, message string, data ...any) {

	if correlationId == "" {
		log.Printf("[conn: %s] ERROR: %s %v", connectionId, message, data)
	} else {
		log.Printf("[conn: %s][%s] ERROR: %s %v", connectionId, correlationId, message, data)
	}
}

func logFatal(connectionId string, correlationId string, message string, data ...any) {

	if correlationId == "" {
		log.Fatalf("[conn: %s] FATAL: %s %v", connectionId, message, data)
	} else {
		log.Fatalf("[conn: %s][%s] FATAL: %s %v", connectionId, correlationId, message, data)
	}
}
