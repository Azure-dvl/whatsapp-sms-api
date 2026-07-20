package handlers

import (
	"encoding/json"
	"main/http/models"
	"net/http"
)

func ForwardHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		resp := models.ForwardResponse{Success: false, Message: "Method not allowed"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if !isClientReady() {
		resp := models.ForwardResponse{Success: false, Message: "WhatsApp client not connected yet. Scan the QR code first."}
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(resp)
		return
	}

	var req models.ForwardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := models.ForwardResponse{Success: false, Message: "Invalid JSON body"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if len(req.Recipients) == 0 {
		resp := models.ForwardResponse{Success: false, Message: "At least one recipient is required"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if req.Message == "" && (req.Image == nil || req.Image.URL == "") {
		resp := models.ForwardResponse{Success: false, Message: "Either message or image is required"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	results := whatAppClient.ForwardMessage(req)

	allSuccess := true
	for _, res := range results {
		if !res.Success {
			allSuccess = false
			break
		}
	}

	status := http.StatusOK
	msg := "Message forwarded to all recipients"
	if !allSuccess {
		status = http.StatusMultiStatus
		msg = "Some forwards failed"
	}

	resp := models.ForwardResponse{
		Success: allSuccess,
		Message: msg,
		Results: results,
	}
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
