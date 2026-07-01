// Mira esto debeeeee de ir en handlers...pero yo como que lo meti en una sola carpeta todo tu ya despues lo organisas con paciencia

package webhook

import (
	"encoding/json"
	"net/http"
)

type WebhookHandler struct {
	Repo *WebhookRepository
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "POST":
		h.register(w, r)
	case "DELETE":
		h.delete(w, r)
	default:
		http.Error(w, `{"success":false,"message":"Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}


func (h *WebhookHandler) register(w http.ResponseWriter, r *http.Request) {
	// POST

	var req WebhookReg
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false, "message": "Invalid JSON body",
		})
		return
	}

	if req.Phone == "" || req.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false, "message": "Phone and URL are required",
		})
		return
	}

	err := h.Repo.Register(r.Context(), req.Phone, req.URL, req.Secret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false, "message": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true, "message": "Webhook registered",
	})
}

func (h *WebhookHandler) delete(w http.ResponseWriter, r *http.Request) {
	// DELETE

	phone := r.URL.Query().Get("phone")
	if phone == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false, "message": "Phone query param is required",
		})
		return
	}

	err := h.Repo.Delete(r.Context(), phone)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Webhook deleted"})
}