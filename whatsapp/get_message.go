package whatsapp

import "main/http/models"

func (w *WhatsAppClient) GetMessageContent(id string) *models.MessageContent {
	if id == "" {
		return nil
	}

	w.mu.RLock()
	fm, ok := w.forwardableMessages[id]
	w.mu.RUnlock()

	if !ok {
		fm = LoadForwardableMessage(id)
		if fm == nil {
			return nil
		}
	}

	mc := &models.MessageContent{ID: id}
	if fm.imageURL != "" {
		mc.Type = "image"
		mc.Caption = fm.imageCaption
	} else {
		mc.Type = "text"
		mc.Text = fm.text
	}
	return mc
}
