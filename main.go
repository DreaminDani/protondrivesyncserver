package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	protonapi "github.com/henrybear327/Proton-API-Bridge"
	"github.com/henrybear327/Proton-API-Bridge/common"
)

// UploadRequest defines the structure for the incoming JSON request
type UploadRequest struct {
	Filename       string `json:"filename"`
	Base64Document string `json:"base64Document"`
}

// Response defines the structure for the JSON response
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	FileID  string `json:"fileID,omitempty"`
}

func main() {
	port := getEnvOrDefault("PORT", "8080")
	http.HandleFunc("/upload", uploadHandler)

	fmt.Printf("Server listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get filename from URL query parameter or use a default name
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		filename = fmt.Sprintf("upload_%s.bin", time.Now().Format("20060102_150405"))
	}

	// Read the raw POST body (the file content)
	fileData, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to read file: "+err.Error())
		return
	}
	defer r.Body.Close()

	// Rest of the ProtonDrive setup...
	protonUsername := os.Getenv("PROTON_USERNAME")
	protonPassword := os.Getenv("PROTON_PASSWORD")
	targetFolderID := os.Getenv("PROTON_DRIVE_FOLDER_ID")

	if protonUsername == "" || protonPassword == "" {
		respondWithError(w, http.StatusInternalServerError, "Proton Drive credentials not configured")
		return
	}

	ctx := context.Background()
	config := common.NewConfigWithDefaultValues()
	config.AppVersion = "web-drive@5.2.0+95291931"

	credentials := &common.FirstLoginCredentialData{
		Username: protonUsername,
		Password: protonPassword,
	}
	config.FirstLoginCredential = credentials

	protonDrive, _, err := protonapi.NewProtonDrive(
		ctx,
		config,
		nil, // authHandler not needed when using FirstLoginCredential
		nil,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to initialize Proton Drive: "+err.Error())
		return
	}
	defer protonDrive.Logout(ctx)

	// Upload the file
	fileID, _, err := protonDrive.UploadFileByReader(
		ctx,
		targetFolderID,
		filename,
		time.Now(),
		bytes.NewReader(fileData),
		0,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to upload file: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "File uploaded successfully to Proton Drive",
		FileID:  fileID,
	})
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	log.Printf("Error: %d - %s", statusCode, message)
	respondWithJSON(w, statusCode, Response{
		Success: false,
		Message: message,
	})
}

func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	jsonResponse, _ := json.Marshal(payload)
	w.Write(jsonResponse)
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
