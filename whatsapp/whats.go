package whatsapp

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"main/http/models"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mdp/qrterminal"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

type MultimediaType string

const (
	MultimediaNone  MultimediaType = ""
	MultimediaImage MultimediaType = "image"
	MultimediaVideo MultimediaType = "video"
	MultimediaAudio MultimediaType = "audio"
	MultimediaDocument MultimediaType = "document"
)

type ReceivedMessage struct {
	ID              string         `json:"id"`
	From            string         `json:"from"`
	FromPN          string         `json:"from_pn,omitempty"`
	Text            string         `json:"text"`
	IsFromMe        bool           `json:"is_from_me"`
	MultimediaType  MultimediaType `json:"multimedia_type,omitempty"`
	MultimediaCaption string       `json:"multimedia_caption,omitempty"`
}

type forwardableMessage struct {
	text               string
	imageURL           string
	imageDirectPath    string
	imageMediaKey      []byte
	imageFileEncSHA256 []byte
	imageFileSHA256    []byte
	imageFileLength    uint64
	imageMimeType      string
	imageCaption       string
	imageJPEGThumbnail []byte
	imageHeight        uint32
	imageWidth         uint32
}

func getMessageFields(v *events.Message) (text string, fm *forwardableMessage) {
	text = v.Message.GetConversation()
	if text == "" && v.Message.GetExtendedTextMessage() != nil {
		text = v.Message.GetExtendedTextMessage().GetText()
	}
	

	if img := v.Message.GetImageMessage(); img != nil {
		fm = &forwardableMessage{
			imageURL:           img.GetURL(),
			imageDirectPath:    img.GetDirectPath(),
			imageMediaKey:      img.GetMediaKey(),
			imageFileEncSHA256: img.GetFileEncSHA256(),
			imageFileSHA256:    img.GetFileSHA256(),
			imageFileLength:    img.GetFileLength(),
			imageMimeType:      img.GetMimetype(),
			imageCaption:       img.GetCaption(),
			imageJPEGThumbnail: img.GetJPEGThumbnail(),
			imageHeight:        img.GetHeight(),
			imageWidth:         img.GetWidth(),
		}
		if text == "" {
			text = img.GetCaption()
		}
	}

	return text, fm
}

func (fm *forwardableMessage) buildMessage() *waE2E.Message {
	msgText := fm.text
	if msgText == "" {
		msgText = fm.imageCaption
	} else if fm.imageCaption != "" {
		msgText = fm.text + "\n" + fm.imageCaption
	}

	if fm.imageURL != "" {
		return &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				URL:           &fm.imageURL,
				DirectPath:    &fm.imageDirectPath,
				MediaKey:      fm.imageMediaKey,
				FileEncSHA256: fm.imageFileEncSHA256,
				FileSHA256:    fm.imageFileSHA256,
				FileLength:    &fm.imageFileLength,
				Mimetype:      proto.String(fm.imageMimeType),
				Caption:       proto.String(msgText),
				JPEGThumbnail: fm.imageJPEGThumbnail,
				Height:        &fm.imageHeight,
				Width:         &fm.imageWidth,
				ContextInfo: &waE2E.ContextInfo{
					IsForwarded:     proto.Bool(true),
					ForwardingScore: proto.Uint32(1),
				},
			},
		}
	} else if fm.text != "" {
		return &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: proto.String(fm.text),
				ContextInfo: &waE2E.ContextInfo{
					IsForwarded:     proto.Bool(true),
					ForwardingScore: proto.Uint32(1),
				},
			},
		}
	}
	return nil
}

func (w *WhatsAppClient) getSenderPN(v *events.Message) (string, string) {
	sender := v.Info.Sender
	fromStr := sender.ToNonAD().String()
	fromPN := ""

	if v.Info.AddressingMode == types.AddressingModeLID {
		if !v.Info.SenderAlt.IsEmpty() {
			fromPN = v.Info.SenderAlt.ToNonAD().String()
		}
	}

	if v.Info.IsGroup && !sender.IsEmpty() {
		return fromStr, fromPN
	}

	target := v.Info.Chat
	return target.ToNonAD().String(), fromPN
}

type WhatsAppClient struct {
	Client *whatsmeow.Client
	Ctx    context.Context

	mu               sync.RWMutex
	receivedMessages  []ReceivedMessage
	forwardableMessages map[string]*forwardableMessage

	Connected chan struct{}
}

func NewWhatsAppClient() *WhatsAppClient {
	return &WhatsAppClient{
		Connected:          make(chan struct{}),
		forwardableMessages: make(map[string]*forwardableMessage),
	}
}

func (w *WhatsAppClient) GetReceivedMessages() []ReceivedMessage {
	w.mu.RLock()
	defer w.mu.RUnlock()
	result := make([]ReceivedMessage, len(w.receivedMessages))
	copy(result, w.receivedMessages)
	return result
}

func (w *WhatsAppClient) handleCommand(v *events.Message, text string) bool {
	if !strings.HasPrefix(text, "/") {
		return false
	}

	var response string
	parts := strings.SplitN(text, " ", 2)
	command := strings.ToLower(parts[0])

	switch command {
	case "/start":
		response = `¡Bienvenido! 🎉

Soy un bot diseñado para ayudarte a promocionar tus productos de venta online. Conmigo podrás reenviar tus publicaciones a múltiples grupos y canales de forma automática.

📌 Usa /new para comenzar a reenviar un mensaje.
📌 Usa /settings para configurar los grupos, canales y horarios.
📌 Usa /help para ver la lista de comandos disponibles.

¡Tu número ha sido registrado! Ahora puedes usar /new para empezar.`
	case "/new":
		response = `Esperando el mensaje que se va a reenviar... 📨

Envíame el mensaje (texto o imagen) que quieres publicar en los grupos y canales configurados.`
	case "/settings":
		response = `⚙️ Configuración disponible:

/forward — Configurar los números, grupos y canales de destino.
/time — Configurar el horario de reenvío (ej: cada día a las 10:00 AM durante 3 días).

Usa /forward o /time para más detalles.`
	case "/help":
		response = `📖 Comandos disponibles:

/start — Registrarse y recibir información del bot.
/new — Enviar un nuevo mensaje para reenviar.
/settings — Ver y configurar opciones de reenvío.
/help — Mostrar esta ayuda.

🔧 Más funciones próximamente.`
	default:
		response = `Comando no reconocido. Usa /help para ver los comandos disponibles.`
	}

	targetJID := v.Info.Chat
	_, err := w.Client.SendMessage(w.Ctx, targetJID, &waE2E.Message{
		Conversation: proto.String(response),
	})
	if err != nil {
		fmt.Printf("Error sending command response: %v\n", err)
	}

	return true
}

func (w *WhatsAppClient) EventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		sender, senderPN := w.getSenderPN(v)
		text, fm := getMessageFields(v)

		if !v.Info.IsFromMe {
			// Check if it's a command
			if w.handleCommand(v, text) {
				return
			}
		}

		multimediaType := MultimediaNone
		if fm != nil && fm.imageURL != "" {
			multimediaType = MultimediaImage
		}

		// Store in received message list
		rm := ReceivedMessage{
			ID:       v.Info.ID,
			From:     sender,
			FromPN:   senderPN,
			Text:     text,
			IsFromMe: v.Info.IsFromMe,
			MultimediaType: multimediaType,
		}
		if fm != nil {
			rm.MultimediaCaption = fm.imageCaption
		}

		w.mu.Lock()
		w.receivedMessages = append(w.receivedMessages, rm)
		// Also store the forwardable message data keyed by ID
		if fm != nil {
			if w.forwardableMessages == nil {
				w.forwardableMessages = make(map[string]*forwardableMessage)
			}
			w.forwardableMessages[v.Info.ID] = fm
		}
		w.mu.Unlock()

		if !v.Info.IsFromMe {
			fmt.Printf("📩 Message from %s: %s\n", sender, text)
		}
	}
}

func (w *WhatsAppClient) ForwardReceivedMessage(id string, recipients []string) []models.ForwardResult {
	results := make([]models.ForwardResult, 0, len(recipients))

	w.mu.RLock()
	fm, ok := w.forwardableMessages[id]
	w.mu.RUnlock()

	// If not found as forwardable, treat as text-only
	if !ok {
		for _, recipient := range recipients {
			text := ""
			for _, rm := range w.receivedMessages {
				if rm.ID == id {
					text = rm.Text
					break
				}
			}
			jid, err := parseJID(recipient)
			if err != nil {
				results = append(results, models.ForwardResult{Recipient: recipient, Success: false, Error: err.Error()})
				continue
			}
			waMessage := &waE2E.Message{
				ExtendedTextMessage: &waE2E.ExtendedTextMessage{
					Text: proto.String(text),
					ContextInfo: &waE2E.ContextInfo{
						IsForwarded:     proto.Bool(true),
						ForwardingScore: proto.Uint32(1),
					},
				},
			}
			_, err = w.Client.SendMessage(w.Ctx, jid, waMessage)
			if err != nil {
				results = append(results, models.ForwardResult{Recipient: recipient, Success: false, Error: err.Error()})
			} else {
				results = append(results, models.ForwardResult{Recipient: recipient, Success: true})
			}
		}
		return results
	}

	// Forward using the original message content
	waMessage := fm.buildMessage()
	if waMessage == nil {
		for _, recipient := range recipients {
			results = append(results, models.ForwardResult{Recipient: recipient, Success: false, Error: "Empty message"})
		}
		return results
	}

	for _, recipient := range recipients {
		jid, err := parseJID(recipient)
		if err != nil {
			results = append(results, models.ForwardResult{Recipient: recipient, Success: false, Error: err.Error()})
			continue
		}
		_, err = w.Client.SendMessage(w.Ctx, jid, waMessage)
		if err != nil {
			results = append(results, models.ForwardResult{Recipient: recipient, Success: false, Error: err.Error()})
		} else {
			results = append(results, models.ForwardResult{Recipient: recipient, Success: true})
		}
	}
	return results
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
