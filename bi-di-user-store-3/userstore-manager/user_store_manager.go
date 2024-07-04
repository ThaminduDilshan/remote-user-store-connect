package main

import (
	"context"
	"fmt"
	"log"
	"math"
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
	numRequests               = 100 // Number of concurrent requests
	maxRetries                = 3
	initialBackoff            = 100 * time.Millisecond
	maxBackoff                = 2 * time.Second
)

func main() {

	// Record start time.
	startTime := time.Now()

	// Set up a connection to the server.
	conn, err := grpc.NewClient(intermediateServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewRemoteUserStoreServiceClient(conn)

	var wg sync.WaitGroup

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// Create a unique request ID for each request
			requestID := fmt.Sprintf("auth-request-%d", i)
			authData, _ := structpb.NewStruct(map[string]interface{}{
				"username": "user1",
				"password": "user1",
			})
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Increase timeout to 10 seconds
			defer cancel()

			correlationID := uuid.New().String() // Generate a UUID for the request ID
			for attempt := 0; attempt < maxRetries; attempt++ {
				r, err := client.InvokeUserStore(ctx, &pb.UserStoreRequest{
					Id:            correlationID,
					OperationType: "DO_AUTHENTICATE",
					Organization:  "test_org_1",
					Data:          authData,
				})
				if err == nil {
					log.Printf("UserStore response for request %s: %s, %s", requestID, r.Id, r.Data.Fields["response"].GetStringValue())
					return
				}

				log.Printf("attempt %d: could not invoke user store for request %s: %v", attempt+1, requestID, err)
				backoff := time.Duration(math.Min(float64(initialBackoff)*(math.Pow(2, float64(attempt))), float64(maxBackoff)))
				time.Sleep(backoff)
			}
			log.Printf("could not invoke user store for request %s after %d attempts", requestID, maxRetries)
		}(i)
	}

	// Wait for all goroutines to finish.
	wg.Wait()

	// Calculate elapsed time and print.
	elapsedTime := time.Since(startTime)
	log.Printf("All requests processed. Took %s to complete.", elapsedTime)
}
