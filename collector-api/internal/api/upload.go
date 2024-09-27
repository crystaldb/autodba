package api

import (
	"collector-api/internal/auth"
	"collector-api/internal/config"
	"io"
	"log"
	"net/http"
	"os"

	"encoding/xml"
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

	// Parse the multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB limit
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	if values, ok := r.MultipartForm.Value["key"]; ok {
		for _, value := range values {
			if cfg.Debug {
				log.Printf("checking for %s: %s\n", "key", value)
			}
			if !auth.S3Authenticate(value) {
				if cfg.Debug {
					log.Printf("Unauthorized access attempt from %s", r.RemoteAddr)
				}
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			break
		}
	} else {
		if cfg.Debug {
			log.Printf("Api Key not found\n")
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if cfg.Debug {
		log.Printf("Authenticated request from %s", r.RemoteAddr)
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
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

	fullPath := filepath.Join(cfg.StorageDir, filename)
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	resp := s3UploadResponse{Key: fullPath}
	responseXML, _ := xml.Marshal(resp)

	w.Header().Set("Content-Type", "application/xml")
	w.Write(responseXML)

	if cfg.Debug {
		log.Printf("Grant response successfully sent to %s", r.RemoteAddr)
	}
}
