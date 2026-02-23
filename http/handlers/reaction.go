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
