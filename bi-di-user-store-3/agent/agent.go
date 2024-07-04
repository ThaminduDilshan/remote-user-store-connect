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
	noOfAgentConnections      = 10
)

func main() {

	log.Println("Starting Local Agent...")
	var wg sync.WaitGroup

	for i := 0; i < noOfAgentConnections; i++ { // Start 5 connections for load handling
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			conn, err := grpc.NewClient(intermediateServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}
			defer conn.Close()

			client := pb.NewUserStoreHubServiceClient(conn)
			stream, err := client.Communicate(context.Background())
			if err != nil {
				log.Fatalf("could not communicate: %v", err)
			}
			defer stream.CloseSend()

			log.Printf("Agent %d: Connection established", i)

			// Send initial connection message with organization ID
			connectMessage := &pb.RemoteMessage{
				OperationType: "CLIENT_CONNECT",
				Id:            uuid.New().String(),
				Organization:  organization,
				Data:          &structpb.Struct{},
			}
			if err := stream.Send(connectMessage); err != nil {
				log.Fatalf("Agent %d: failed to send connect message: %v", i, err)
			}

			for {
				req, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						log.Printf("Agent %d: Stream closed by server", i)
						return
					}
					log.Fatalf("Agent %d: failed to receive: %v", i, err)
				}
				log.Printf("Agent %d: Received request: %s with data: %v", i, req.Id, req.Data)

				var response string
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
				if err := stream.Send(&pb.RemoteMessage{Id: req.Id, OperationType: req.OperationType, Organization: req.Organization, Data: responseData}); err != nil {
					log.Fatalf("Agent %d: failed to send response: %v", i, err)
				}
				log.Printf("Agent %d: Sent response for request: %s", i, req.Id)
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
