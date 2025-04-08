package main


import (
	"context"
	"github.com/miguel-m7/whatsapp-service/server/manage"
	"go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/store/sqlstore"  // Adicione esta linha
    waLog "go.mau.fi/whatsmeow/util/log"  // Adicione esta linha também
	"fmt"
	"time"
	"log"
	"go.mau.fi/whatsmeow/types"
	"net"
	// "stream"
	"google.golang.org/grpc"
	pb "github.com/miguel-m7/whatsapp-service/pb"
	_ "github.com/mattn/go-sqlite3"
	// "strconv"
)


type server struct {
	pb.UnimplementedWhatsAppServiceServer // Embed the unimplemented server for forward compatibility
	container *sqlstore.Container
	clients *[]whatsmeow.Client	

}



func (s *server) StartSession(req *pb.StartSessionRequest, stream pb.WhatsAppService_StartSessionServer)  error {
    log.Println("Start Session...")
	device, err := manage.CreateClient(s.container)
	if err != nil {
		log.Println("Error creating client:", err)
		return err
	}
	qrChan, err := manage.GetQRChannel(device)
	if err != nil {
		log.Println("Error getting QR channel:", err)
		return err
	}
	err = device.Connect()
	if err != nil {
		log.Println("Error connecting client:", err)
	}
	nextID := len(*s.clients)
	log.Println("Next ID:", device.Store.LID)
	for evt := range qrChan {
		if evt.Event == "code" {
			// Render the QR code here
			// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
			fmt.Println("QR code:", evt.Code)
			select {
				case <-stream.Context().Done():
					return stream.Context().Err()
				default:
					response := &pb.SessionResponse{
						SessionId: int64(nextID),
						QrCode:   evt.Code,
						Status:   "QR Code generated",
					}
					if err := stream.Send(response); err != nil {
						return err
					}
					fmt.Printf("Sent Session ID: %s, QR Code: %s\n", int64(nextID), evt.Code)
			}
		} else {
			fmt.Println("Login event:", evt.Event)
			*s.clients = append(*s.clients, *device)
			if device.Store.ID != nil {
				response := &pb.SessionResponse{
					SessionId: int64(nextID),
					QrCode:   evt.Code,
					Status:   "Logged",
					Jid:   device.Store.ID.String(),
				}
				if err := stream.Send(response); err != nil {
					return err
				}
			} else {
				response := &pb.SessionResponse{
					SessionId: int64(nextID),
					QrCode:   evt.Code,
					Status:   "Failed",
				}
				if err := stream.Send(response); err != nil {
					return err
				}
			}
			
			log.Println("Next ID:", device.Store.ID)
	}
	}
    return nil
}

func (s *server) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	// Implement your method logic here
	jid, err := types.ParseJID(req.SenderJid)
	if err != nil {
		log.Println("Error parsing JID:", err)
		return &pb.SendMessageResponse{
			Success: false,
			MessageId: "",
			Status: "Invalid JID",
		}, err
	}
	cli := manage.GetClientByJID(s.container, jid)
	err = manage.SendMessage(cli, req.ToPhoneNumber, req.MessageContent)
	if err != nil {
		log.Println("Error sending message:", err)
		return  &pb.SendMessageResponse{
			Success: false,
			MessageId: "",
			Status: "Failed to send message",
		}, err
	}
	return &pb.SendMessageResponse{
		Success: true,
		MessageId: "67890",
		Status: "Message sent successfully",
	}, nil
}


func (s *server) CheckSessionStatus(ctx context.Context, req *pb.SessionStatusRequest) (*pb.SessionStatusResponse, error) {
	t , err := manage.CheckSessionByJID(s.container, req.Jid, 10*time.Second)
	log.Println("Session status:", t.IsLoggedIn())
	if err != nil {
		log.Println("Error checking session status:", err)
		return &pb.SessionStatusResponse{
			IsConnected: false,
			QrCode: "",
		}, nil
	}
	return &pb.SessionStatusResponse{
		IsConnected: true,
		QrCode: "",
	}, nil
}



func main() {
	
	dbLog := waLog.Stdout("Database", "INFO", false)
	container, err := sqlstore.New("sqlite3", "file:./data/mainstore.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}



	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	pb.RegisterWhatsAppServiceServer(s, &server{
		container: container,
		clients: &[]whatsmeow.Client{},
	})
	// Register the server with gRPC
	log.Println("Server is running on port 50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}



}