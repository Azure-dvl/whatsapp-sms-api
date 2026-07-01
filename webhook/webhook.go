// Ejem bueno aca te deje toda la talla del manejo
// Despues organiza esta talla x carpeta

package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type WebhookDispatcher struct {
	repo   *WebhookRepository
	client *http.Client
}

func NewWebhookDispatcher(repo *WebhookRepository) *WebhookDispatcher {
	return &WebhookDispatcher{
		repo:   repo,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (d *WebhookDispatcher) Dispatch(ctx context.Context, evt WebhookPay) {
	fmt.Println("entering to Dispatch")
	go func() {
		fmt.Println("A")
		phone := extractPhone(evt.From)
		entry, err := d.repo.GetByPhone(ctx, phone)
		fmt.Println("B")

		if err != nil || entry == nil {
			return
		}

		fmt.Println("Sending the dispatch to retry function")
		d.sendWithRetry(entry, evt)
	}()
	fmt.Println("outgoing from Dispatch")
}

// 3 intentos tops de fallo
func (d *WebhookDispatcher) sendWithRetry(entry *WebhookEntry, payload WebhookPay) {
	body, _ := json.Marshal(payload)
	signature := computeHMAC(body, entry.Secret)

	for attempt := 0; attempt < 3; attempt++ {
		req, _ := http.NewRequest("POST", entry.URL, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Webhook-Signature", "sha256="+signature)
		req.Header.Set("X-Webhook-Event", payload.Event)

		resp, err := d.client.Do(req)
		fmt.Println("Request sent")
		if err == nil && resp.StatusCode < 300 {
			resp.Body.Close()
			return
		}

		if attempt < 2 {
			time.Sleep(time.Duration(1<<attempt) * time.Second)
		}
	}
}

func extractPhone(jid string) string {
	idx := strings.IndexByte(jid, '@')
	if idx == -1 {
		return jid
	}
	return jid[:idx]
}

func computeHMAC(data []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}
