package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/henrybear327/Proton-API-Bridge/common"
)

// RequestPayload defines the expected structure of the incoming JSON request
type RequestPayload struct {
	Filename string `json:"filename"`
	Data     string `json:"data"`
}

// Authenticate and initialize Proton Drive
func initializeProtonDrive() (*ProtonAPIBridge.ProtonDrive, error) {
	ctx := context.Background()
	config := &common.Config{
		Username: "your-username",
		Password: "your-password",
	}
	authHandler := ProtonAPIBridge.NewDefaultAuthHandler()

	protonDrive, _, err := ProtonAPIBridge.NewProtonDrive(ctx, config, authHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Proton Drive: %v", err)
	}

	return protonDrive, nil
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var payload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	decodedData, err := base64.StdEncoding.DecodeString(payload.Data)
	if err != nil {
		http.Error(w, "Failed to decode base64 data", http.StatusBadRequest)
		return
	}

	tempFilePath := "/tmp/" + payload.Filename
	if err := ioutil.WriteFile(tempFilePath, decodedData, 0644); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFilePath) // Clean up the temp file after upload

	// Initialize Proton Drive
	protonDrive, err := initializeProtonDrive()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to initialize Proton Drive: %v", err), http.StatusInternalServerError)
		return
	}

	// Upload the file
	err = protonDrive.UploadFile(tempFilePath, payload.Filename)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload to Proton Drive: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))
}

func main() {
	http.HandleFunc("/upload", uploadHandler)
	fmt.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", nil)
}
