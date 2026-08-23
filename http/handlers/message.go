package handlers

import (
	"encoding/json"
	"main/http/models"
	"net/http"
)

func MessageHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		resp := models.MessageResponse{Success: false, Message: "Method not allowed"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(resp)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		resp := models.MessageResponse{Success: false, Message: "id query parameter is required"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	content := whatAppClient.GetMessageContent(id)
	if content == nil {
		resp := models.MessageResponse{Success: false, Message: "Message not found"}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := models.MessageResponse{
		Success: true,
		Message: "Message found",
		Data:    content,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
