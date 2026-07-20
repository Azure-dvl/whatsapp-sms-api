package handlers

import "main/whatsapp"

var whatAppClient *whatsapp.WhatsAppClient

func SetClient(client *whatsapp.WhatsAppClient) {
	whatAppClient = client
}

func isClientReady() bool {
	if whatAppClient == nil || whatAppClient.Client == nil {
		return false
	}
	select {
	case <-whatAppClient.Connected:
		return true
	default:
		return false
	}
}
