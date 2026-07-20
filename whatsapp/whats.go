package whatsapp

import (
	"context"
	"fmt"
	"os"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mdp/qrterminal"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type ReceivedMessage struct {
	ID   string `json:"id"`
	From string `json:"from"`
	Text string `json:"text"`
}

func getSenderPN(v *events.Message) string {
	sender := v.Info.Sender
	// If it's a group message, sender is the person who sent it
	if v.Info.IsGroup && !sender.IsEmpty() {
		if v.Info.AddressingMode == types.AddressingModeLID && !v.Info.SenderAlt.IsEmpty() {
			return v.Info.SenderAlt.ToNonAD().String()
		}
		return sender.ToNonAD().String()
	}
	// For DMs, the sender IS the chat
	target := v.Info.Chat
	if v.Info.AddressingMode == types.AddressingModeLID && !v.Info.SenderAlt.IsEmpty() {
		return v.Info.SenderAlt.ToNonAD().String()
	}
	return target.ToNonAD().String()
}

type WhatsAppClient struct {
	Client *whatsmeow.Client
	Ctx    context.Context

	mu              sync.RWMutex
	receivedMessages []ReceivedMessage

	Connected chan struct{}
}

func NewWhatsAppClient() *WhatsAppClient {
	return &WhatsAppClient{
		Connected: make(chan struct{}),
	}
}

func (w *WhatsAppClient) GetReceivedMessages() []ReceivedMessage {
	w.mu.RLock()
	defer w.mu.RUnlock()
	result := make([]ReceivedMessage, len(w.receivedMessages))
	copy(result, w.receivedMessages)
	return result
}

func (w *WhatsAppClient) EventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		sender := getSenderPN(v)
		text := v.Message.GetConversation()
		if text == "" && v.Message.GetExtendedTextMessage() != nil {
			text = v.Message.GetExtendedTextMessage().GetText()
		}
		if text == "" {
			text = "[Non-text message]"
		}

		w.mu.Lock()
		w.receivedMessages = append(w.receivedMessages, ReceivedMessage{
			ID:   v.Info.ID,
			From: sender,
			Text: text,
		})
		w.mu.Unlock()

		fmt.Printf("📩 Message from %s: %s\n", sender, text)
	}
}

func (w *WhatsAppClient) Connect() {

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	w.Ctx = context.Background()
	container, err := sqlstore.New(w.Ctx, "pgx", os.Getenv("DATABASE_URL"), dbLog)
	if err != nil {
		panic(err)
	}
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice(w.Ctx)
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "INFO", true)
	w.Client = whatsmeow.NewClient(deviceStore, clientLog)
	w.Client.AddEventHandler(w.EventHandler)

	if w.Client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := w.Client.GetQRChannel(context.Background())
		err = w.Client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
				fmt.Println("QR code:", evt.Code)
				fmt.Println("QR code recibido, generando imagen...")
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				err := qrcode.WriteFile(evt.Code, qrcode.Medium, 256, "whatsapp-qr.png")
				if err != nil {
					fmt.Println("Error generando QR:", err)
				} else {
					fmt.Println("QR guardado como whatsapp-qr.png")
				}
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
		close(w.Connected)
	} else {
		// Already logged in, just connect
		err = w.Client.Connect()
		if err != nil {
			panic(err)
		}
		close(w.Connected)
	}
}

func (w *WhatsAppClient) Disconnect() {
	// Disconnect the client when done
	w.Client.Disconnect()
}
