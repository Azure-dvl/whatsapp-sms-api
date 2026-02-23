package whatsapp

import (
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

var lastMessageId types.MessageID

func (w *WhatsAppClient) SendMessage(number string, message string) error {
	jid := types.NewJID(number, types.DefaultUserServer)

	waMessage := &waE2E.Message{
		Conversation: proto.String(message),
	}

	msg, err := w.Client.SendMessage(w.Ctx, jid, waMessage)
	lastMessageId = msg.ID
	return err
}

func (w *WhatsAppClient) SendReaction(number string, reaction string) error {
	target := types.NewJID(number, types.DefaultUserServer)
	message := w.Client.BuildReaction(target, w.Client.Store.GetJID(), lastMessageId, reaction)
	_, err := w.Client.SendMessage(w.Ctx, target, message)
	return err
}
