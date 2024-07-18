package main

import (
	"context"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"bi-di-user-store-3/agent/config"
	"bi-di-user-store-3/agent/sec_vault"
	pb "bi-di-user-store-3/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

func main() {

	log.Println("Starting Local Agent...")
	var wg sync.WaitGroup

	log.Println("Reading configurations.")
	config := config.ReadConfig("deployment.toml")

	// Read the AES key from environment variables.
	encSecretKey, keyExists := os.LookupEnv("SECRET_KEY")
	if keyExists && encSecretKey != "" {
		log.Println("Secret key found in the environment variables.")
	}

	hubServiceAddress := config.HubService.Host + ":" + strconv.Itoa(config.HubService.Port)

	for i := 0; i < config.System.NoOfIdleConnections; i++ { // Start 5 connections for load handling
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			agentId := uuid.New().String()
			log.Printf("Agent conn %s: Starting connection with the server", agentId)

			conn, err := grpc.NewClient(hubServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}
			defer conn.Close()

			client := pb.NewUserStoreHubServiceClient(conn)

			// Authentication.
			agentToken := config.Security.Token

			if agentToken == "" {
				log.Fatalf("Agent conn %s: Agent token not found", agentId)
			} else if strings.Contains(agentToken, sec_vault.SECRET_KEY_PREFIX) {
				agentToken, err = sec_vault.ResolveSecret(agentToken, config, encSecretKey)

				if err != nil {
					log.Fatalf("Agent conn %s: Failed to resolve secret: %v", agentId, err)
				}
			}

			md := metadata.Pairs("authorization", "Bearer "+agentToken)
			ctx := metadata.NewOutgoingContext(context.Background(), md)

			stream, err := client.Communicate(ctx)
			if err != nil {
				log.Fatalf("could not communicate: %v", err)
			}
			defer stream.CloseSend()

			log.Printf("Agent conn %s: Connection initialized with the server", agentId)

			// Send initial connection message with tenant ID.
			connectMessage := &pb.RemoteMessage{
				OperationType: "CLIENT_CONNECT",
				RequestId:     agentId,
				Tenant:        config.System.Tenant,
				UserStore:     config.System.UserStore,
				Data:          &structpb.Struct{},
			}
			if err := stream.Send(connectMessage); err != nil {
				log.Fatalf("Agent conn %s: failed to send connect message: %v", agentId, err)
			}

			for {
				req, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						log.Printf("Agent conn %s: Stream closed by server", agentId)
						return
					}
					log.Fatalf("Agent conn %s: failed to receive: %v", agentId, err)
				}

				var response string
				if req.OperationType == "SERVER_CONNECTED" {
					log.Printf("Agent conn %s: Successfully connected with the server: %s", agentId, req.RequestId)
					continue
				}

				log.Printf("Agent conn %s: Received request: %s with data: %v", agentId, req.RequestId, req.Data)

				if req.OperationType == "DO_AUTHENTICATE" {
					username := req.Data.Fields["username"].GetStringValue()
					password := req.Data.Fields["password"].GetStringValue()
					response = processAuthenticationRequest(username, password)
				} else {
					message := req.Data.Fields["message"].GetStringValue()
					response = processUserRequest(message)
				}

				responseData, _ := structpb.NewStruct(map[string]interface{}{
					"response": response,
				})
				if err := stream.Send(&pb.RemoteMessage{
					CorrelationId: req.CorrelationId,
					RequestId:     req.RequestId,
					OperationType: req.OperationType,
					Tenant:        config.System.Tenant,
					UserStore:     config.System.UserStore,
					Data:          responseData,
				}); err != nil {
					log.Fatalf("Agent conn %s: failed to send response: %v", agentId, err)
				}
				log.Printf("Agent conn %s: Sent response for request: %s", agentId, req.RequestId)
			}
		}(i)
	}

	wg.Wait()
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
