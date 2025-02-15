package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
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
	FileID  string `json:"fileID,omitempty"` // Optionally include FileID on success
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

	// 1. Decode JSON request
	var uploadRequest UploadRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&uploadRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	if uploadRequest.Filename == "" || uploadRequest.Base64Document == "" {
		respondWithError(w, http.StatusBadRequest, "Filename and base64Document are required")
		return
	}

	// 2. Decode base64 document
	decodedDocument, err := base64.StdEncoding.DecodeString(uploadRequest.Base64Document)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid base64 document: "+err.Error())
		return
	}

	// 3. Retrieve credentials from environment variables
	protonUsername := os.Getenv("PROTON_USERNAME")
	protonPassword := os.Getenv("PROTON_PASSWORD")
	targetFolderID := os.Getenv("PROTON_DRIVE_FOLDER_ID") // Optional folder ID

	if protonUsername == "" || protonPassword == "" {
		respondWithError(w, http.StatusInternalServerError, "Proton Drive credentials not configured (PROTON_USERNAME, PROTON_PASSWORD)")
		return
	}

	ctx := context.Background()
	config := common.NewConfigWithDefaultValues()
	config.Username = protonUsername
	config.Password = protonPassword

	_, _, cred, _, err := common.Login(ctx, config, nil, nil) // Using common.Login for authentication
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Proton Drive authentication failed: "+err.Error())
		return
	}

	protonDrive, _, err := protonapi.NewProtonDrive(ctx, config, nil, cred.AccessToken, cred.RefreshToken) // Create ProtonDrive instance
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to initialize Proton Drive client: "+err.Error())
		return
	}

	// 5. Upload to Proton Drive
	var parentLinkID string = "" // Root folder by default
	if targetFolderID != "" {
		parentLinkID = targetFolderID
	}

	fileReader := bytes.NewReader(decodedDocument)
	fileID, _, err := protonDrive.UploadFileByReader(ctx, parentLinkID, uploadRequest.Filename, time.Now(), fileReader, len(decodedDocument), nil)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Proton Drive upload failed: "+err.Error())
		return
	}

	// 6. Respond with success
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
	jsonResponse, _ := json.Marshal(payload) // Error ignored for simplicity in example
	w.Write(jsonResponse)
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
