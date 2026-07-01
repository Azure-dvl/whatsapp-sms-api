// Aca estan los manejos de los eventos
// Por ahora solo mensajes

package webhook

import (
	"context"
	"fmt"

	"go.mau.fi/whatsmeow/types/events"
)

func HandleIncomingMessage(dispatcher *WebhookDispatcher, evt *events.Message) {
	info := evt.Info
	msg := evt.Message

	payload := WebhookPay{
		Event:     "message",
		Timestamp: info.Timestamp,
		From:      info.Sender.ToNonAD().String(),
		FromJID:   info.Sender.String(),
	}

	payload.Data = TextMessage{Type: "text", Text: *msg.Conversation, ID: info.ID}

	fmt.Println("Dispatching")
	dispatcher.Dispatch(context.Background(), payload)
}
