package handlers

import (
	"encoding/json"
	"net/http"
	"os"
)

// Ruta donde se guarda el QR (debe coincidir con la usada en whatsapp.Connect)
const qrFilePath = "whatsapp-qr.png"

func QRHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !isClientReady() {
		// Check if there's a QR file on disk (first time setup)
		if _, err := os.Stat(qrFilePath); err == nil {
			w.Header().Set("Content-Type", "image/png")
			data, _ := os.ReadFile(qrFilePath)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
			return
		}
		resp := map[string]interface{}{"success": false, "message": "WhatsApp client not connected yet. Scan the QR code first."}
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Verificar si el archivo existe
	if _, err := os.Stat(qrFilePath); os.IsNotExist(err) {
		http.Error(w, "QR no disponible", http.StatusNotFound)
		return
	}

	// Leer el archivo
	data, err := os.ReadFile(qrFilePath)
	if err != nil {
		http.Error(w, "Error leyendo QR", http.StatusInternalServerError)
		return
	}

	// Servir la imagen
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
