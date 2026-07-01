package http

import (
	"fmt"
	"log"
	"main/http/handlers"
	"main/webhook"
	"main/whatsapp"
	"net/http"
	"os"
)

func SetupHandlers(client *whatsapp.WhatsAppClient, webhookHandler *webhook.WebhookHandler) {
	handlers.SetClient(client)
	http.Handle("/sms", http.HandlerFunc(handlers.SmsHandler))
	http.Handle("/reaction", http.HandlerFunc(handlers.ReactionHandler))
	http.Handle("/qr", http.HandlerFunc(handlers.QRHandler))
	
	// El webhook
	http.Handle("/webhook", webhookHandler)
}

func Serve() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "9050"
	}
	address := fmt.Sprintf("0.0.0.0:%v", port)
	log.Default().Printf("Starting server on %s", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
