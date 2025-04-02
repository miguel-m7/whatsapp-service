package server


import (
	"log"
	"net"
	"google.golang.org/grpc"
	"pb"
)

type server struct {}



func (s *server) StartSession(ctx context.Context, req *pb.YourRequest) (*pb.YourResponse, error) {
	// Implement your method logic here
	return &pb.SessionResponse{}, nil
}


func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	pb.RegisterYourServiceServer(s, &server{}) // Register your service here
	log.Println("Server is running on port 50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}



}