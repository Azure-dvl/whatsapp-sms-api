package handlers

import (
	"encoding/json"
	"main/http/models"
	"net/http"
)

func GroupsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		resp := models.GroupsResponse{Success: false}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if !isClientReady() {
		resp := models.GroupsResponse{Success: false}
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(resp)
		return
	}

	groups, err := whatAppClient.GetGroupsAndNewsletters()
	if err != nil {
		resp := models.GroupsResponse{Success: false}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := models.GroupsResponse{
		Success: true,
		Groups:  groups,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
