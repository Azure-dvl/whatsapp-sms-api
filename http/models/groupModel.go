package models

type GroupItem struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	JID   string `json:"jid"`
	Type  string `json:"type"`
}

type GroupsResponse struct {
	Success bool        `json:"success"`
	Groups  []GroupItem `json:"groups"`
}
