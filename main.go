package main

import (
	"database/sql"
	"fmt"
	"main/http"
	"main/webhook"
	"main/whatsapp"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	fmt.Println("Starting connect the WhatsappApi")
	whatsAppClient := &whatsapp.WhatsAppClient{}
	go whatsAppClient.Connect()
	fmt.Println("WhatsappApi connected successfully")

	// Webhook
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL")) // lo puse por env y no tengo como probar esta talla ahora mismo so te toca testear esto o hosteralo para yo testearlo...Tu primero revisa bien el codigo.
	if err != nil{
		panic(err)
	}
	defer db.Close()

	repo := webhook.NewWebhookRepository(db)
	repo.CreateTable(nil)

	dispatcher := webhook.NewWebhookDispatcher(repo)
	whatsAppClient.Dispatcher = dispatcher

	webhook := &webhook.WebhookHandler{Repo: repo}

	http.SetupHandlers(whatsAppClient, webhook)
	fmt.Println("HTTP handlers set up successfully")
	http.Serve()
}
