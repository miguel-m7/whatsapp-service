package manage

import (
	"context"
	"log"
	"fmt"
	"time"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	// "google.golang.org/protobuf/proto"
	waProto "go.mau.fi/whatsmeow/binary/proto"
)

// CreateClient creates a new WhatsApp client with a new device
func CreateClient(container *sqlstore.Container) (*whatsmeow.Client, error) {
	deviceStore := container.NewDevice()
	if deviceStore == nil {
		return nil, fmt.Errorf("failed to create device")
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	// err := client.Connect()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to connect: %v", err)
	// }
	return client, nil
}

// GetQRChannel returns a QR channel for new device connection
func GetQRChannel(client *whatsmeow.Client) (<-chan whatsmeow.QRChannelItem, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	qrChan, err := client.GetQRChannel(context.Background())
	// err := client.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	return qrChan, err
}

// SendMessage sends a message to a specific JID
func SendMessage(client *whatsmeow.Client, jid string, message string) error {
	if client == nil {
		return fmt.Errorf("client is nil")
	}
	recipient, err := types.ParseJID(jid)
	if err != nil {
		return err
	}
	msg := &waProto.Message{
		Conversation: &message,
	}
	_, err = client.SendMessage(context.Background(), recipient, msg)
	return err
}

// GetClientByJID retrieves a client by JID from the store
func GetClientByJID(container *sqlstore.Container, jid types.JID) *whatsmeow.Client {
	device, err := container.GetDevice(jid)
	if err != nil {
		return nil
	}
	if device == nil {
		return nil
	}
	clientLog := waLog.Stdout("Client", "INFO", false)
	client := whatsmeow.NewClient(device, clientLog)
	err = client.Connect()
	if err != nil {
		return nil
	}
	return client
}



// CheckSessionByJID verifica a sessão do cliente WhatsApp pelo JID fornecido
// CheckSessionByJID verifica a sessão do cliente WhatsApp pelo JID fornecido
func CheckSessionByJID(container *sqlstore.Container, jid string, timeout time.Duration) (*whatsmeow.Client, error) {
	if container == nil {
		return nil, fmt.Errorf("container is nil")
	}

	// Parse o JID fornecido
	parsedJID, err := types.ParseJID(jid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JID: %v", err)
	}

	// Tente obter o dispositivo associado ao JID
	device, err := container.GetDevice(parsedJID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device for JID %s: %v", jid, err)
	}
	if device == nil {
		return nil, fmt.Errorf("no device found for JID %s", jid)
	}

	// Crie um cliente WhatsApp utilizando o dispositivo recuperado
	clientLog := waLog.Stdout("Client", "INFO", false)
	client := whatsmeow.NewClient(device, clientLog)

	// Crie um contexto com timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Canal para sinalizar a conclusão da autenticação
	authChannel := make(chan bool, 1)

	// Função para conectar e autenticar o cliente
	go func() {
		err := client.Connect()
		time.Sleep(2 * time.Second) // Aguarde 2 segundos antes de verificar a autenticação
		if err != nil {
			log.Println("Failed to connect: %v", err)
			authChannel <- false
			return
		}
		if client.IsLoggedIn() {
			log.Println("Successfully authenticated")
			authChannel <- true
		} else {
			log.Println("Authentication failed")
			authChannel <- false
		}
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("authentication timeout after %v", timeout)
	case authSuccess := <-authChannel:
		if authSuccess {
			return client, nil
		}
		return nil, fmt.Errorf("client for JID %s is not authenticated", jid)
	}
}
