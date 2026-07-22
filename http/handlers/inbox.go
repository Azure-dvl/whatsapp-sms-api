package handlers

import (
	"encoding/json"
	"net/http"
)

func InboxHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		resp := map[string]interface{}{"success": false, "message": "Method not allowed"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if !isClientReady() {
		resp := map[string]interface{}{"success": false, "message": "WhatsApp client not connected yet. Scan the QR code first."}
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(resp)
		return
	}

	messages := whatAppClient.GetReceivedMessages()
	resp := map[string]interface{}{
		"success":  true,
		"messages": messages,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
