package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

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

func addMemo(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
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

		log.Printf("[%s] SERVER [REQUEST] [METHOD: %s] Received new memo with title: %s\n", time.Now().Format(time.RFC3339), r.Method, newMemo.Title)

		response, err := json.Marshal(newMemo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(response)
	} else if r.Method == http.MethodGet {
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
	router := mux.NewRouter()
	router.HandleFunc("/note", addMemo).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/note/{id}", addMemo).Methods(http.MethodGet, http.MethodDelete, http.MethodPatch)

	fmt.Println("Server is running on port 8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}

