package models

type SendMessageRequest struct {
	Phone   string `json:"phone,omitempty"`
	Chat    string `json:"chat,omitempty"`
	Message string `json:"message"`
}

type SendMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
