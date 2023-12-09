package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"strconv"
	"strings"
)

type Configuration struct {
	ServicePort int      `json:"servicePort"`
	Sync        string   `json:"sync"`
	Replicas    []string `json:"replicas"`
}

type Memo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

var primaryAddress string
var localMemos map[int]Memo

type Acknowledgement struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func logRequest(r *http.Request, args ...interface{}) {
    message := fmt.Sprint(args...)
    fmt.Printf("[%s] Replica SERVER [REQUEST] [METHOD: %s] %s\n", time.Now().Format(time.RFC3339), r.Method, message)
}

func forwardWriteRequest(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	req, err := http.NewRequest(r.Method, primaryAddress+r.URL.Path, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Forwarding %s request to primary server: %s\n", r.Method, req.URL.String())

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = updatePrimaryMemo(body)
	if err != nil {
		fmt.Println("Failed to update primary server:", err)
		http.Error(w, "Failed to update primary server", http.StatusInternalServerError)
		return
	}

	for key, val := range resp.Header {
		w.Header().Set(key, val[0])
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

func updatePrimaryMemo(body []byte) error {
	var ack Acknowledgement
	err := json.Unmarshal(body, &ack)
	if err != nil {
		return err
	}

	if !ack.Success {
		return fmt.Errorf("primary server failed to update memo: %s", ack.Message)
	}

	return nil
}

func handleReadRequest(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")

	if len(pathParts) == 3 && pathParts[2] != "" {
		id, err := strconv.Atoi(pathParts[2])
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		if memo, ok := localMemos[id]; ok {
			response, err := json.Marshal(memo)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(response)
			return
		}

		http.Error(w, "Memo not found locally", http.StatusNotFound)
		return
	}

	if len(localMemos) == 0 {
		http.Error(w, "No Data", http.StatusNotFound)
		return
	}

	memosSlice := make([]Memo, 0, len(localMemos))
	for _, v := range localMemos {
		memosSlice = append(memosSlice, v)
	}

	response, err := json.Marshal(memosSlice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response)
}

func main() {
	localMemos = make(map[int]Memo)

	if len(os.Args) != 2 {
		fmt.Println("Usage: go run %s config.json", os.Args[0])
		return
	}

	configFile := os.Args[1]

	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error reading the config file: %s\n", err)
	}

	var config Configuration
	err = json.Unmarshal(configData, &config)
	if err != nil {
		fmt.Printf("Error decoding the config JSON: %s\n", err)
	}

	if len(config.Replicas) == 0 {
		fmt.Printf("No replicas found in config")
	}

	// Primary server is assumed to be the first replica
	primaryAddress = config.Replicas[0]

	http.HandleFunc("/note/", func(w http.ResponseWriter, r *http.Request) {
		logRequest(r, "print");
		switch r.Method {
		case http.MethodGet:
			handleReadRequest(w, r)
		case http.MethodPost, http.MethodDelete, http.MethodPatch:
			forwardWriteRequest(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Replica Server is running...")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		fmt.Println("%s", err)
	}
}

