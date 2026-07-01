// Vale te hago docs para que sepas esta talla y veas si lo hago bien
// Sigo un poco perdida la verdad asi q voy a hacer lo q vi en google y la guia extensa nivel dios que me dio DeepSeek y OpenCode
// No se GO asi q la sintaxis va a estar cruda ok? y muy copiada de lo que tu has hecho xq ver un video en yt es perdida de tiempo de corriente

package webhook

import "time"


type WebhookReg struct {
	Phone	string	`json:"phone"`
	URL		string	`json:"url"`
	Secret	string	`json:"secret"`
}

type WebhookEntry struct {
	ID		int		`json:"id"`
	Phone	string	`json:"phone"`
	URL		string	`json:"url"`
    Secret    string    `json:"-"`
	CreatedAt	time.Time	`json:"created_at"`
}

// Payload que se envia al webhook
type WebhookPay struct {
    Event     string      `json:"event"`
    Timestamp time.Time   `json:"timestamp"`
    From      string      `json:"from"`
    FromJID   string      `json:"fromJid"`
    Data      interface{} `json:"data"`
}


// Estos son los tipos de eventos (Solo texto x ahora xq no se que mas quieras meter o que mas hay...no me da tiempo a leer todo esto solo lo importante (OpenCode se dio a la tarea de resumirme toda la estructura que tenias en esto y entendi pero no fue tan especifico))

const (
    EventMessage       = "message"
    EventReceipt       = "receipt"
    EventPresence      = "presence"
)

type TextMessage struct {
    Type string `json:"type"`
    Text string `json:"text"`
    ID   string `json:"id"`
}
