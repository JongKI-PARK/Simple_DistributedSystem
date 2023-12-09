package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
	"io/ioutil"
	"os"
	"path/filepath"
	"github.com/gorilla/mux"
)

type Configuration struct {
	ServicePort int		`json:"servicePort"`
	Sync		string	`json:"sync"`
	Replicas	[]string`json:"replicas"`
}

type Memo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

var (
	memos   []Memo
	memosMu sync.Mutex
	idCount = 0
)

func logRequest(r *http.Request, args ...interface{}) {
    message := fmt.Sprint(args...)
    fmt.Printf("[%s] Replica SERVER [REQUEST] [METHOD: %s] %s\n", time.Now().Format(time.RFC3339), r.Method, message)
}

func addMemo(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		//logRequest(r, "Received POST request")

		var newMemo Memo
		err := json.NewDecoder(r.Body).Decode(&newMemo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		memosMu.Lock()
		defer memosMu.Unlock()

		idCount++
		newMemo.ID = idCount
		memos = append(memos, newMemo)

		logRequest(r, "Received new memo with title: ", newMemo.Title)

		response, err := json.Marshal(newMemo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(response)
	} else if r.Method == http.MethodGet {
		logRequest(r, "Received GET request")

		params := mux.Vars(r)
		if idStr, ok := params["id"]; ok {
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}

			memosMu.Lock()
			defer memosMu.Unlock()

			for _, memo := range memos {
				if memo.ID == id {
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
			}

			http.Error(w, "Memo not found", http.StatusNotFound)
			return
		}

		memosMu.Lock()
		defer memosMu.Unlock()

		if len(memos) == 0 {
			http.Error(w, "No data", http.StatusNotFound)
			return
		}

		response, err := json.Marshal(memos)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(response)
	} else if r.Method == http.MethodDelete {
		logRequest(r, "Received DELETE request")

		params := mux.Vars(r)
		if idStr, ok := params["id"]; ok {
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}

			memosMu.Lock()
			defer memosMu.Unlock()

			for i, memo := range memos {
				if memo.ID == id {
					memos = append(memos[:i], memos[i+1:]...)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"msg": "OK"}`))
					return
				}
			}

			http.Error(w, "Memo not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Invalid endpoint", http.StatusBadRequest)
	} else if r.Method == http.MethodPatch {
		logRequest(r, "Received PATCH request")

		params := mux.Vars(r)
		if idStr, ok := params["id"]; ok {
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}

			var requestBody map[string]string
			err = json.NewDecoder(r.Body).Decode(&requestBody)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			memosMu.Lock()
			defer memosMu.Unlock()

			for i, memo := range memos {
				if memo.ID == id {
					// Update the memo with the new body if provided in the request
					if newBody, ok := requestBody["body"]; ok {
						memos[i].Body = newBody
					}

					if newTitle, ok := requestBody["title"]; ok {
						memos[i].Title = newTitle
					}

					response, err := json.Marshal(memos[i])
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(response)
					return
				}
			}

			http.Error(w, "Memo not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Invalid endpoint", http.StatusBadRequest)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage : go run %s config.json\n", filepath.Base(os.Args[0]))
		return
	}

	configFile := os.Args[1]
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error reading the config file: %s\n", err)
	}

	var config Configuration
	err = json.Unmarshal(configData, &config)
	if err != nil {
		log.Fatalf("Error decoding the config JSON: %s\n", err)
	}

	fmt.Printf("Service Port: %d\n", config.ServicePort)
	fmt.Printf("Sync Method: %s\n", config.Sync)
	fmt.Println("Replicas:")
	for _, replica := range config.Replicas {
		fmt.Println(replica)
	}

	router := mux.NewRouter()
	router.HandleFunc("/note", addMemo).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/note/{id}", addMemo).Methods(http.MethodGet, http.MethodDelete, http.MethodPatch)

	fmt.Println("Replica Server is running on port 8081...")
	if err := http.ListenAndServe(":8081", router); err != nil {
		log.Fatal(err)
	}
}

