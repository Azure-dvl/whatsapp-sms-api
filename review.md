# Análisis Completo del Proyecto `whatsapp-sms-api`

> Fecha del análisis: 2026-06-30

---

## 1. Estructura del Proyecto

```
whatsapp-sms-api/
├── .dockerignore
├── .env.example
├── .gitignore
├── Dockerfile
├── go.mod
├── go.sum
├── LICENSE
├── main.go
├── http/
│   ├── http.go
│   └── handlers/
│   │   ├── client.go
│   │   ├── qr.go
│   │   ├── reaction.go
│   │   └── sms.go
│   └── models/
│       ├── reactionModel.go
│       └── smsModel.go
├── whatsapp/
│   ├── sendMessage.go
│   └── whats.go
```

---

## 2. Archivos y Contenido

### `main.go`
```go
package main

import (
	"fmt"
	"main/http"
	"main/whatsapp"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	fmt.Println("Starting connect the WhatsappApi")
	whatsAppClient := &whatsapp.WhatsAppClient{}
	go whatsAppClient.Connect()
	fmt.Println("WhatsappApi connected successfully")

	http.SetupHandlers(whatsAppClient)
	fmt.Println("HTTP handlers set up successfully")
	http.Serve()
}
```

### `http/http.go`
```go
package http

import (
	"fmt"
	"log"
	"main/http/handlers"
	"main/whatsapp"
	"net/http"
	"os"
)

func SetupHandlers(client *whatsapp.WhatsAppClient) {
	handlers.SetClient(client)
	http.Handle("/sms", http.HandlerFunc(handlers.SmsHandler))
	http.Handle("/reaction", http.HandlerFunc(handlers.ReactionHandler))
	http.Handle("/qr", http.HandlerFunc(handlers.QRHandler))
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
```

### `http/models/smsModel.go`
```go
package models

type SendMessageRequest struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
}

type SendMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
```

### `http/models/reactionModel.go`
```go
package models

type SendReactionRequest struct {
	Phone    string `json:"phone"`
	Reaction string `json:"reaction"`
}
```

### `http/handlers/client.go`
```go
package handlers

import "main/whatsapp"

var whatAppClient *whatsapp.WhatsAppClient

func SetClient(client *whatsapp.WhatsAppClient) {
	whatAppClient = client
}
```

### `http/handlers/sms.go`
```go
package handlers

import (
	"encoding/json"
	"main/http/models"
	"net/http"
)

func SmsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		resp := models.SendMessageResponse{Success: false, Message: "Method not allowed"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if whatAppClient == nil {
		resp := models.SendMessageResponse{Success: false, Message: "WhatsApp client not initialized"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	var req models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := models.SendMessageResponse{Success: false, Message: "Invalid JSON body"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if req.Phone == "" || req.Message == "" {
		resp := models.SendMessageResponse{Success: false, Message: "Phone and message are required"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	err := whatAppClient.SendMessage(req.Phone, req.Message)
	if err != nil {
		resp := models.SendMessageResponse{Success: false, Message: err.Error()}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := models.SendMessageResponse{Success: true, Message: "Message sent successfully"}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
```

### `http/handlers/qr.go`
```go
package handlers

import (
	"net/http"
	"os"
)

// Ruta donde se guarda el QR (debe coincidir con la usada en whatsapp.Connect)
const qrFilePath = "whatsapp-qr.png"

func QRHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(qrFilePath); os.IsNotExist(err) {
		http.Error(w, "QR no disponible", http.StatusNotFound)
		return
	}

	data, err := os.ReadFile(qrFilePath)
	if err != nil {
		http.Error(w, "Error leyendo QR", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
```

### `http/handlers/reaction.go`
```go
package handlers

import (
	"encoding/json"
	"fmt"
	"main/http/models"
	"net/http"
)

func sendReaction(req models.SendReactionRequest) error {
	err := whatAppClient.SendReaction(req.Phone, req.Reaction)
	return err
}

func deleteReaction() {

}

func ReactionHandler(w http.ResponseWriter, r *http.Request) {

	if whatAppClient == nil {
		resp := models.SendMessageResponse{Success: false, Message: "WhatsApp client not initialized"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	var req models.SendReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := models.SendMessageResponse{Success: false, Message: "Invalid JSON body"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if req.Phone == "" {
		resp := models.SendMessageResponse{Success: false, Message: "Phone is required"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if req.Reaction == "" {
		req.Reaction = "❤"
	}

	switch r.Method {
	case "POST":
		err := sendReaction(req)
		var resp models.SendMessageResponse
		if err != nil {
			resp = models.SendMessageResponse{Success: false, Message: fmt.Sprintf("Error sending reaction: %v", err)}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp = models.SendMessageResponse{
			Success: true,
			Message: "Reaction sent successfuly",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return
	case "DELETE":
		resp := models.SendMessageResponse{Success: false, Message: "Not implemented yet"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	default:
		resp := models.SendMessageResponse{Success: false, Message: "Method not allowed"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(resp)
		return
	}
}
```

### `whatsapp/sendMessage.go`
```go
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
```

### `whatsapp/whats.go`
```go
package whatsapp

import (
	"context"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mdp/qrterminal"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type WhatsAppClient struct {
	Client *whatsmeow.Client
	Ctx    context.Context
}

func (w *WhatsAppClient) Connect() {

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	w.Ctx = context.Background()
	container, err := sqlstore.New(w.Ctx, "pgx", os.Getenv("DATABASE_URL"), dbLog)
	if err != nil {
		panic(err)
	}
	deviceStore, err := container.GetFirstDevice(w.Ctx)
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "INFO", true)
	w.Client = whatsmeow.NewClient(deviceStore, clientLog)
	// w.Client.AddEventHandler(EventHandler)

	if w.Client.Store.ID == nil {
		qrChan, _ := w.Client.GetQRChannel(context.Background())
		err = w.Client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
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
	} else {
		err = w.Client.Connect()
		if err != nil {
			panic(err)
		}
	}
}

func (w *WhatsAppClient) Disconnect() {
	w.Client.Disconnect()
}

// func EventHandler(evt interface{}) {
// 	switch v := evt.(type) {
// 	case *events.Message:
// 		fmt.Println("Received a message!", v.Message.GetConversation())
// 	}
// }
```

---

## 3. Hallazgos Detallados

### 3.1 TODOs, FIXMEs, HACKs y Código Incompleto

| # | Archivo | Línea | Tipo | Descripción |
|---|---------|-------|------|-------------|
| 1 | `whatsapp/whats.go` | 36 | Código comentado | `// w.Client.AddEventHandler(EventHandler)` -- Manejador de eventos deshabilitado |
| 2 | `whatsapp/whats.go` | 77-82 | Código comentado | Función `EventHandler` completa comentada -- No se procesan mensajes entrantes |
| 3 | `http/handlers/reaction.go` | 15-17 | Función vacía | `deleteReaction()` es un stub sin implementación |
| 4 | `http/handlers/reaction.go` | 64-68 | No implementado | Caso DELETE retorna `"Not implemented yet"` |
| 5 | `http/handlers/reaction.go` | 10-13 | Código redundante | `sendReaction` es un wrapper de una línea que no agrega valor |

### 3.2 Bugs Potenciales

| # | Archivo | Línea | Severidad | Descripción |
|---|---------|-------|-----------|-------------|
| 1 | `whatsapp/sendMessage.go` | 9, 19, 25 | **Crítico** | **Race condition**: `lastMessageId` es una variable global (`var lastMessageId types.MessageID`). `SendMessage` escribe en ella (línea 19) y `SendReaction` la lee (línea 25). En concurrencia hay data race. Además, `SendReaction` siempre reacciona al último mensaje enviado globalmente, no a uno específico. |
| 2 | `main.go` | 15-16 | **Crítico** | `whatsAppClient.Connect()` se lanza en goroutine (line 15) e inmediatamente se imprime `"WhatsappApi connected successfully"` (line 16) **antes de que la conexión esté establecida**. |
| 3 | `main.go` | 18-20 | **Crítico** | `http.SetupHandlers(whatsAppClient)` y `http.Serve()` se ejecutan sin esperar a que WhatsApp esté conectado. El servidor HTTP sirve peticiones antes de que el cliente WhatsApp esté listo. |
| 4 | `whatsapp/sendMessage.go` | 18-19 | **Alto** | **Posible nil pointer dereference**: Si `w.Client.SendMessage()` falla, `msg` podría ser `nil`, pero se accede a `msg.ID` en la línea 19 antes de verificar `err`. |
| 5 | `main.go` | 12 | **Alto** | `godotenv.Load()` -- El error retornado se ignora. Si no hay archivo `.env`, no hay advertencia. |
| 6 | `http/http.go` | 25 | **Medio** | `http.ListenAndServe(address, nil)` -- Sin `ReadTimeout`, `WriteTimeout` ni `IdleTimeout`. Vulnerable a resource exhaustion por conexiones lentas. |
| 7 | `http/handlers/qr.go` | - | **Bajo** | No se cierra `r.Body.Close()` en el handler QR. |
| 8 | `whatsapp/whats.go` | 40 | **Medio** | `qrChan, _ := w.Client.GetQRChannel(context.Background())` -- El segundo valor de retorno (error) se ignora con `_`. |

### 3.3 Funcionalidades Faltantes

| # | Funcionalidad | Impacto |
|---|---------------|---------|
| 1 | **Manejo de mensajes entrantes** | La API solo puede enviar mensajes, no recibirlos ni responder automáticamente |
| 2 | **ID de mensaje en reacciones** | No se puede especificar a qué mensaje reaccionar; siempre usa `lastMessageId` global |
| 3 | **Soporte para mensajes multimedia** | Solo texto plano. Sin imágenes, documentos, audio o video |
| 4 | **Mensajes a grupos** | Solo JIDs de usuario individual. Sin soporte para grupos |
| 5 | **Endpoints de health check** | Sin `/health` o `/ready` para verificar estado del servicio |
| 6 | **Autenticación/Autorización** | API completamente abierta. Sin API keys, tokens ni autenticación |
| 7 | **Rate limiting** | Sin protección contra abuso o DoS |
| 8 | **Graceful shutdown** | No se capturan señales SIGINT/SIGTERM. El servidor muere abruptamente |
| 9 | **Reconexión automática** | Si la conexión WhatsApp se pierde, no hay lógica de reconexión |
| 10 | **Logging de peticiones** | Sin request ID, sin logging de requests entrantes |
| 11 | **Validación de números telefónicos** | No se valida formato E.164, código de país, etc. |
| 12 | **WebSocket/SSE para QR** | No hay mecanismo push; el cliente debe pollear `/qr` |

### 3.4 Problemas de Calidad de Código

| # | Archivo | Línea | Problema |
|---|---------|-------|----------|
| 1 | `go.mod` | 1 | El módulo se llama `main`, lo cual es inusual y puede causar confusión. Convención: `github.com/user/repo` |
| 2 | Varios | - | Comentarios mezclan inglés y español sin consistencia |
| 3 | `main.go:13` | - | `"Starting connect the WhatsappApi"` -- Error gramatical ("connect" debería ser "connecting") |
| 4 | `http/handlers/reaction.go:59` | - | Typo: `"successfuly"` debería ser `"successfully"` |
| 5 | `http/handlers/client.go` | - | Variable global `whatAppClient` dificulta el testing |
| 6 | Varios | - | Sin interfaces: todo es con tipos concretos, imposible mockear |
| 7 | `whatsapp/sendMessage.go` | - | No hay wrapping de errores con contexto |
| 8 | `http/handlers/qr.go:9` | - | Ruta del QR hardcodeada; debería ser configurable |
| 9 | `http/handlers/reaction.go:10-13` | - | `sendReaction()` es un wrapper innecesario |
| 10 | Varios | - | `json.NewEncoder(w).Encode(resp)` -- Errores de encoding ignorados en múltiples lugares |
| 11 | `whatsapp/whats.go:23,34` | - | Niveles de log hardcodeados (`"DEBUG"`, `"INFO"`) |

### 3.5 Problemas de Seguridad

| # | Archivo | Severidad | Descripción |
|---|---------|-----------|-------------|
| 1 | N/A | **Crítico** | Sin autenticación ni autorización -- cualquiera con acceso a la red puede enviar mensajes |
| 2 | `http/handlers/sms.go` | **Alto** | Los errores del cliente WhatsApp (incluyendo errores internos de DB) se exponen directamente al cliente vía `err.Error()` |
| 3 | `whatsapp/whats.go` | **Alto** | `panic(err)` en producción -- cualquier error de DB o conexión mata el proceso |
| 4 | N/A | **Alto** | Sin TLS/HTTPS -- todo el tráfico viaja en texto plano |
| 5 | `http/handlers/sms.go` | **Medio** | Sin sanitización de entrada -- phone y message se pasan directamente al cliente WhatsApp |
| 6 | N/A | **Medio** | Sin secrets management -- `DATABASE_URL` solo en variable de entorno |
| 7 | `Dockerfile` | **Bajo** | Sin `EXPOSE` -- no documenta el puerto del contenedor |

### 3.6 Problemas de Dependencias

| # | Problema | Severidad |
|---|----------|-----------|
| 1 | **Go 1.25.0 en `go.mod`** -- Esta versión de Go no existe. Ningún compilador actual puede compilarlo. | **Crítico** |
| 2 | `google.golang.org/protobuf` listada sin comentario `// indirect` | Bajo |
| 3 | `github.com/mattn/go-sqlite3` incluida como dependencia indirecta aunque el proyecto usa PostgreSQL | Bajo |

### 3.7 Problemas de Configuración

| # | Archivo | Problema |
|---|---------|----------|
| 1 | `.env.example` | Solo documenta `DATABASE_URL` -- falta `PORT` que sí se usa en `http.go` |
| 2 | `whatsapp/whats.go` | No hay valor por defecto para `DATABASE_URL`; si no está configurada, `sqlstore.New` falla con panic |
| 3 | `http/handlers/qr.go` y `whatsapp/whats.go` | Ruta del QR hardcodeada en dos archivos (`"whatsapp-qr.png"`). Si cambia uno sin el otro, se desincronizan |
| 4 | N/A | Sin distinción entre entorno dev/prod |
| 5 | `Dockerfile` | No expone puerto, no tiene multi-stage build |

### 3.8 Valores Hardcodeados

| # | Archivo | Línea | Valor | Debería ser |
|---|---------|-------|-------|-------------|
| 1 | `http/http.go` | 22 | `"9050"` | Configurable (parcialmente, vía `PORT`, pero con fallback hardcodeado) |
| 2 | `http/http.go` | 24 | `"0.0.0.0"` | Configurable vía variable de entorno |
| 3 | `http/handlers/qr.go` | 9 | `"whatsapp-qr.png"` | Configurable vía variable de entorno |
| 4 | `whatsapp/whats.go` | 53 | `"whatsapp-qr.png"` | Configurable (debe coincidir con el de `qr.go`) |
| 5 | `whatsapp/whats.go` | 53 | `256` | Configurable (tamaño del QR) |
| 6 | `whatsapp/whats.go` | 53 | `qrcode.Medium` | Configurable (nivel de recuperación del QR) |
| 7 | `whatsapp/whats.go` | 23 | `"DEBUG"` | Configurable (nivel de log DB) |
| 8 | `whatsapp/whats.go` | 34 | `"INFO"` | Configurable (nivel de log cliente) |
| 9 | `http/handlers/reaction.go` | 44 | `"❤"` | Configurable (reacción por defecto) |

### 3.9 Validación y Manejo de Errores Faltante

| # | Archivo | Línea | Descripción |
|---|---------|-------|-------------|
| 1 | `main.go` | 12 | `godotenv.Load()` -- error ignorado |
| 2 | `main.go` | 15 | Error de `Connect()` no propagado (panic dentro de goroutine) |
| 3 | `whatsapp/whats.go` | 40 | Error de `GetQRChannel` ignorado |
| 4 | `http/handlers/qr.go` | 28 | `w.Write(data)` -- retorno ignorado |
| 5 | `http/handlers/sms.go` | varios | `json.NewEncoder(w).Encode(resp)` -- errores ignorados |
| 6 | `http/handlers/reaction.go` | varios | `json.NewEncoder(w).Encode(resp)` -- errores ignorados |
| 7 | `http/handlers/sms.go` | - | No se valida formato del número de teléfono |

### 3.10 Pruebas Ausentes

- No existe ningún archivo `*_test.go` en todo el proyecto.
- Cobertura de pruebas: **0%**.
- No hay tests unitarios, de integración ni end-to-end.

### 3.11 Inconsistencias y Convenciones

| # | Descripción |
|---|-------------|
| 1 | Variable `whatAppClient` (con 'A' minúscula) vs tipo `WhatsAppClient` (con 'A' mayúscula). Además lleva doble 'p' innecesaria |
| 2 | Comentarios en español (`qr.go`) e inglés (`whats.go`) mezclados |
| 3 | Nombres de archivos: `smsModel.go`, `reactionModel.go` -- el sufijo `Model` es redundante (el package `models` ya indica su propósito) |
| 4 | `sendMessage.go` usa camelCase; convención Go es `send_message.go` |
| 5 | `sms.go` usa `r.Method != http.MethodPost` (constante), `reaction.go` usa `case "POST"` (string literal) -- inconsistente |
| 6 | `reaction.go` usa `models.SendMessageResponse` para respuestas de reacción -- semánticamente incorrecto |
| 7 | `deleteReaction()` declarada pero nunca llamada -- el case DELETE retorna "Not implemented" sin invocarla |

---

## 4. Resumen de Prioridades

### Crítico (arreglar inmediatamente)
1. **Go 1.25.0 no existe** -- Cambiar a una versión de Go real (ej. 1.22 o 1.23)
2. **Race condition en `lastMessageId`** -- Mover a ámbito de instancia o usar mutex
3. **Server HTTP sirve antes de que WhatsApp esté listo** -- Agregar sincronización (WaitGroup, channel, etc.)
4. **Posible nil pointer en `msg.ID`** -- Verificar `err` antes de acceder a `msg`
5. **Sin graceful shutdown** -- Capturar SIGINT/SIGTERM
6. **`panic()` en producción** -- Reemplazar con manejo de errores adecuado

### Alto
1. **Sin autenticación** -- Agregar API key o token
2. **Exposición de errores internos** -- No retornar errores crudos al cliente
3. **Mensajes entrantes deshabilitados** -- Reactivar EventHandler o documentar la limitación
4. **Sin health checks** -- Agregar endpoint `/health`
5. **Sin reconexión** -- Implementar lógica de reconexión

### Medio
1. **Sin timeouts HTTP** -- Configurar ReadTimeout, WriteTimeout, IdleTimeout
2. **Validación de entrada** -- Validar phone (E.164), message (longitud)
3. **Rate limiting**
4. **Valores hardcodeados → configurables por env**
5. **Typo en `reaction.go`** (successfuly)

### Bajo
1. **Consistencia idioma comentarios** -- Unificar a inglés
2. **Renombrar módulo `main` → URL estándar**
3. **`r.Body.Close()` faltante en QR handler**
4. **Eliminar `deleteReaction()` o implementarlo**
5. **Eliminar `sendReaction()` wrapper innecesario**
6. **Agregar tests**
7. **Agregar `EXPOSE` en Dockerfile**

---

## 5. Conclusión

El proyecto es un MVP funcional que expone una API HTTP para enviar mensajes y reacciones de WhatsApp usando la librería `whatsmeow` con almacenamiento PostgreSQL. Sin embargo, tiene problemas **críticos** que impiden su compilación (Go 1.25.0) y su operación en producción (race conditions, panics, falta de graceful shutdown). La seguridad es inexistente (API abierta, sin TLS). Se recomienda una refactorización significativa antes de considerar su uso en producción.
