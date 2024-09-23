package api

import (
	// "collector-api/internal/auth"
	"collector-api/internal/config"
	"log"
	"net/http"

	"encoding/xml"
	"io/ioutil"
	"path/filepath"
)

type s3UploadResponse struct {
	Location string
	Bucket   string
	Key      string
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfigWithDefaultPath()
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		if cfg.Debug {
			log.Printf("Error loading config: %v", err)
		}
		return
	}

	// Authenticate the request
	// if !auth.Authenticate(r) {
	// 	if cfg.Debug {
	// 		log.Printf("Unauthorized access attempt from %s", r.RemoteAddr)
	// 	}
	// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 	return
	// }

	// if cfg.Debug {
	// 	log.Printf("Authenticated request from %s", r.RemoteAddr)
	// }

	// Parse the multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB limit
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusInternalServerError)
		return
	}

	filename := fileHeader.Filename
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	if cfg.Debug {
		log.Printf("Received file [%s] with size: %d bytes\n", filename, len(data))
	}

	filePath := filepath.Join(cfg.StorageDir, filename)
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	resp := s3UploadResponse{Key: filename}
	responseXML, _ := xml.Marshal(resp)

	w.Header().Set("Content-Type", "application/xml")
	w.Write(responseXML)

	if cfg.Debug {
		log.Printf("Grant response successfully sent to %s", r.RemoteAddr)
	}
}
