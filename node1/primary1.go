package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
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
}

func main() {
	http.HandleFunc("/note", addMemo)

	fmt.Println("Server is running on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

