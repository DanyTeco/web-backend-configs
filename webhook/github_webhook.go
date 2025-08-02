package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type GitHubWebhook struct {
	Repository struct {
		Name     string `json:"name"`
		CloneURL string `json:"clone_url"`
	} `json:"repository"`
}

const (
	logFile       = "/var/log/webhook-deploy.log"
	webhookSecret = "xxxxxxxxxxxxxxxx" // Replace with your actual secret
	timeout       = 30 * time.Second
)

func writeLog(message string) {
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Cannot open log file: %v", err)
		return
	}
	defer f.Close()
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	f.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, message))
}

func verifySignature(signature string, body []byte) bool {
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)
	expectedSignature := "sha256=" + hex.EncodeToString(expectedMAC)
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

func runDeployScriptAsync(project, cloneURL string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "/bin/bash", "/path/to/deploy.sh", project, cloneURL)
		output, err := cmd.CombinedOutput()

		writeLog(fmt.Sprintf("Running deploy script for project: %s", project))
		writeLog(string(output))

		if err != nil {
			writeLog(fmt.Sprintf("Deployment error: %v", err))
		} else {
			writeLog(fmt.Sprintf("Deployment success for project: %s", project))
		}
	}()
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		http.Error(w, "Missing signature", http.StatusUnauthorized)
		writeLog("Missing signature")
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Cannot read body", http.StatusBadRequest)
		return
	}

	if !verifySignature(signature, body) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		writeLog("Invalid signature")
		return
	}

	var payload GitHubWebhook
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	project := strings.ToLower(payload.Repository.Name)
	writeLog(fmt.Sprintf("Webhook triggered for project: %s", project))
	runDeployScriptAsync(project, payload.Repository.CloneURL)

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Deployment started for project: %s", project)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("OK"))
}

func main() {
	log.Println("Starting Webhook Deploy Server on :6666...")
	writeLog("Starting Webhook Deploy Server...")

	r := mux.NewRouter()
	r.HandleFunc("/webhook", webhookHandler).Methods("POST")
	r.HandleFunc("/health", healthHandler).Methods("GET")

	srv := &http.Server{
		Handler:      r,
		Addr:         ":6666",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
