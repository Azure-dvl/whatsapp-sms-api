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

	if !isClientReady() {
		resp := models.SendMessageResponse{Success: false, Message: "WhatsApp client not connected yet. Scan the QR code first."}
		w.WriteHeader(http.StatusServiceUnavailable)
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

	if req.Chat == "" && req.Phone == "" {
		resp := models.SendMessageResponse{Success: false, Message: "Phone or chat is required"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if req.Message == "" {
		resp := models.SendMessageResponse{Success: false, Message: "Message is required"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	var err error
	if req.Chat != "" {
		err = whatAppClient.SendMessageToJID(req.Chat, req.Message)
	} else {
		err = whatAppClient.SendMessage(req.Phone, req.Message)
	}

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
