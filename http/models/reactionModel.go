package models

type SendReactionRequest struct {
	Phone    string `json:"phone"`
	Reaction string `json:"reaction"`
}
