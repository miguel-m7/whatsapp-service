package main 

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"time"
	// "fmt"
	"io"
	pb "github.com/miguel-m7/whatsapp-service/pb"


)


func main() {
	connection , err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer connection.Close()

	client := pb.NewWhatsAppServiceClient(connection)
	startSession(client)
}



func startSession(client pb.WhatsAppServiceClient) {
    request := &pb.StartSessionRequest{}
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    stream, err := client.StartSession(ctx, request)
    if err != nil {
        log.Fatalf("Error while calling StartSession: %v", err)
    }

    for {
        response, err := stream.Recv()
        if err == io.EOF {
            // End of stream
            break
        }
        if err != nil {
            log.Fatalf("Error while receiving: %v", err)
        }

        log.Printf("Received QR Code: %s", response.QrCode)
    }
}