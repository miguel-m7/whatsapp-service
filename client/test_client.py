import unittest
from unittest.mock import Mock, patch, MagicMock
import grpc
from client.client import grpc_channel, start_session, main
import whatsapp_service_pb2 as pb

# filepath: d:\projects\wwpservice\client\test_client.py

class TestWhatsAppClient(unittest.TestCase):
    def test_grpc_channel_creation(self):
        with patch('grpc.insecure_channel') as mock_channel:
            mock_instance = Mock()
            mock_channel.return_value = mock_instance
            
            with grpc_channel("test:50051"):
                mock_channel.assert_called_once_with("test:50051")
            
            mock_instance.close.assert_called_once()

    def test_start_session_success(self):
        mock_client = Mock()
        mock_response = Mock()
        mock_response.qr_code = "test_qr_code"
        mock_stream = MagicMock()
        mock_stream.__enter__.return_value = iter([mock_response])
        mock_client.StartSession.return_value = mock_stream

        with patch('logging.info') as mock_logging:
            start_session(mock_client)
            mock_logging.assert_called_with("Received QR Code: test_qr_code")

    def test_start_session_error(self):
        mock_client = Mock()
        mock_client.StartSession.side_effect = grpc.RpcError()

        with patch('logging.error') as mock_logging:
            start_session(mock_client)
            mock_logging.assert_called_once()

    def test_main(self):
        mock_client = Mock()
        mock_channel = Mock()
        
        with patch('client.client.grpc_channel') as mock_grpc_channel, \
             patch('whatsapp_service_pb2_grpc.WhatsAppServiceStub') as mock_stub:
            
            mock_grpc_channel.return_value.__enter__.return_value = mock_channel
            mock_stub.return_value = mock_client
            
            main()
            
            mock_stub.assert_called_once_with(mock_channel)
            mock_client.StartSession.assert_called_once()

if __name__ == '__main__':
    unittest.main()