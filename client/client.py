import grpc
import logging
import pb.whatsapp_pb2 as pb
import pb.whatsapp_pb2_grpc as pb_grpc

# Set up logging
logging.basicConfig(level=logging.INFO)

class WhatsAppClient:
    """A reusable gRPC client for WhatsAppService."""

    def __init__(self, target: str):
        """Initialize the client with the gRPC server target."""
        self.target = target
        self.channel = None
        self.stub = None

    def connect(self):
        """Establish a connection to the gRPC server."""
        self.channel = grpc.insecure_channel(self.target)
        self.stub = pb_grpc.WhatsAppServiceStub(self.channel)

    def disconnect(self):
        """Close the connection to the gRPC server."""
        if self.channel:
            self.channel.close()
            self.channel = None
            self.stub = None
    
    def SendMessage(self, request: pb.SendMessageRequest) -> pb.SendMessageResponse:
        """Send a message to a specified phone number."""
        if not self.stub:
            raise RuntimeError("Client is not connected. Call connect() first.")
        try:
            response = self.stub.SendMessage(request)
            return response
        except grpc.RpcError as e:
            logging.error(f"Error during SendMessage: {e}")
            raise

    def CheckSessionStatus(self, request: pb.SessionStatusRequest) -> pb.SessionStatusResponse:
        """Check the status of the WhatsApp session."""
        if not self.stub:
            raise RuntimeError("Client is not connected. Call connect() first.")
        try:
            response = self.stub.CheckSessionStatus(request)
            return response
        except grpc.RpcError as e:
            logging.error(f"Error during CheckSessionStatus: {e}")
            raise

    def start_session(self):
        """Start a session and yield QR codes from the server."""
        if not self.stub:
            raise RuntimeError("Client is not connected. Call connect() first.")
        request = pb.StartSessionRequest()
        try:
            stream = self.stub.StartSession(request)
            for response in stream:
                yield response
        except grpc.RpcError as e:
            logging.error(f"Error during StartSession: {e}")
            raise

    def __enter__(self):
        """Enable usage as a context manager."""
        self.connect()
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        """Clean up resources when exiting the context."""
        self.disconnect()



if __name__ == "__main__":
    # Example usage
    import qrcode
    target = "localhost:50051"
    user_isAuth = True
    jid = '556194487140:11@s.whatsapp.net'
    send_to_number = '557193473101@s.whatsapp.net'
    with WhatsAppClient(target) as client:
        try:
            # check session status
            resp = client.CheckSessionStatus(
                pb.SessionStatusRequest(jid=jid)
            )
            logging.info(f"Session status: {resp.is_connected}")
            # for resp in client.start_session():
            #     img = qrcode.make(resp.qr_code)
            #     img.save("qr_code.png")
            #     logging.info("QR Code saved as qr_code.png")
            #     logging.info(f'Response: {resp}')
            #     logging.info(f'Received Status: {resp.status}')
            #     logging.info(f'Received JID: {resp.jid}')
            #     logging.info(f'Received SessionId: {resp.session_id}')
            #     if resp.jid and resp.status == 'Logged':
            #         user_isAuth = True
            #         jid = resp.jid
            # if resp.is_connected:
            #     logging.info("User is authenticated.")
            #     logging.info("Try send message")
            #     resp = client.SendMessage(
            #         pb.SendMessageRequest(
            #             sender_jid=jid,
            #             to_phone_number=send_to_number,
            #             message_content="Mensagem T???este"
            #         )
            #     )
            #     print(resp)
            

        except Exception as e:
            logging.error(f"An error occurred: {e}")