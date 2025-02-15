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

	// Read the raw file data from the request body
	fileData, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to read request body: "+err.Error())
		return
	}
	defer r.Body.Close()

	log.Printf("Received file data of size: %d bytes", len(fileData))

	// Generate filename with timestamp
	filename := fmt.Sprintf("upload_%s.txt", time.Now().Format("20060102_150405"))

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
	config.AppVersion = "macos-drive@1.0.0-alpha.1+rclone"

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

	log.Printf("Successfully uploaded file '%s' with ID: %s", filename, fileID)
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
