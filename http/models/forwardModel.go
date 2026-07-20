package models

type ForwardImage struct {
	URL      string `json:"url"`
	MimeType string `json:"mime_type,omitempty"`
	Caption  string `json:"caption,omitempty"`
}

type ForwardRequest struct {
	Message    string         `json:"message,omitempty"`
	Image      *ForwardImage  `json:"image,omitempty"`
	Recipients []string       `json:"recipients"`
}

type ForwardResult struct {
	Recipient string `json:"recipient"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

type ForwardResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Results []ForwardResult `json:"results,omitempty"`
}
