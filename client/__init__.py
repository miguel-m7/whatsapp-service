from .client import WhatsAppClient, logging

if __name__ == "__main__":
    # Example usage
    target = "localhost:50051"
    with WhatsAppClient(target) as client:
        try:
            while qr_code in client.start_session():
                logging.info(f"Received QR Code: {qr_code}")
        except Exception as e:
            logging.error(f"An error occurred: {e}")