/*
	2023.12.01 ~ 2023.12.12
	Jongki Park

	replica.go : Replica Server code for Distributed System Remote-Write
*/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
	"bytes"
	"io"
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
    fmt.Printf("[%s] Replica SERVER [REQUEST] [METHOD: %s] %s\n", time.Now().Format(time.StampNano), r.Method, message)
}

func getPrimaryURL() (string, error) {
	configFile := os.Args[1]

	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return "", err
	}

	var config Configuration
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return "", err
	}

	if len(config.Replicas) == 0 {
		return "", fmt.Errorf("Invalid config.json file")
	}

	replicaURL := "http://" + config.Replicas[0] + "/note"

	return replicaURL, nil
}

func forwardMemo(w http.ResponseWriter, r *http.Request) {
	primaryURL, err := getPrimaryURL()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    //fmt.Printf("Primary server URL: %s\n", primaryURL)

	if r.Method == http.MethodGet {
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
	} else if r.Method == http.MethodPost {
		fmt.Printf("[%s] Replica SERVER [FORWARD] [METHOD : %s] to [%s]\n", time.Now().Format(time.StampNano), r.Method, primaryURL)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		r.Body = ioutil.NopCloser(bytes.NewReader(body))
		url := primaryURL

		req, err := http.NewRequest(r.Method, url, bytes.NewReader(body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for h, val := range r.Header {
			req.Header[h] = val
		}

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for h, val := range resp.Header {
			w.Header().Set(h, val[0])
		}

		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("Response Status: %s\n", resp.Status)
		bodyContent, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Response Body: %s\n", bodyContent)

	} else if r.Method == http.MethodDelete {
		params := mux.Vars(r)
		idStr, ok := params["id"]
		if !ok {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		primaryURL, err := getPrimaryURL()
		if err != nil {
			http.Error(w, "Failed to get primary URL", http.StatusInternalServerError)
			return
		}

		forwardURL := fmt.Sprintf("%s/%d", primaryURL, id)

		client := http.Client{}
		primaryReq, err := http.NewRequest(http.MethodDelete, forwardURL, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("[%s] Replica SERVER [FORWARD] [METHOD : %s] to [%s]\n", time.Now().Format(time.StampNano), r.Method, forwardURL)

		resp, err := client.Do(primaryReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copy the response headers to the current response writer
		for h, val := range resp.Header {
			w.Header().Set(h, val[0])
		}

		// Write the status code and body from the response
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Print the received response details
		fmt.Printf("Response Status: %s\n", resp.Status)
		bodyContent, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Response Body: %s\n", bodyContent)

	} else if r.Method == http.MethodPatch {
		params := mux.Vars(r)
		idStr, ok := params["id"]
		if !ok {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		requestBody := make(map[string]string)
		err = json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		primaryURL, err := getPrimaryURL()
		if err != nil {
			http.Error(w, "Failed to get primary URL", http.StatusInternalServerError)
			return
		}

		forwardURL := fmt.Sprintf("%s/%d", primaryURL, id)

		client := http.Client{}

		// Prepare the request body
		newBody, bodyExist := requestBody["body"]
		newTitle, titleExist := requestBody["title"]

		var patchData map[string]interface{}
		if bodyExist || titleExist {
			patchData = make(map[string]interface{})
			if bodyExist {
				patchData["body"] = newBody
			}
			if titleExist {
				patchData["title"] = newTitle
			}
		}

		// Convert the data to JSON
		requestData, err := json.Marshal(patchData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		primaryReq, err := http.NewRequest(http.MethodPatch, forwardURL, bytes.NewReader(requestData))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Copy headers from the original request to the new request
		for h, val := range r.Header {
			primaryReq.Header[h] = val
		}

		fmt.Printf("[%s] Replica SERVER [FORWARD] [METHOD : %s] to [%s]\n", time.Now().Format(time.StampNano), r.Method, forwardURL)
		resp, err := client.Do(primaryReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copy the response headers to the current response writer
		for h, val := range resp.Header {
			w.Header()[h] = val
		}

		// Write the status code and body from the response
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Print the received response details
		fmt.Printf("Response Status: %s\n", resp.Status)
		bodyContent, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Response Body: %s\n", bodyContent)

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
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

func requestFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fromPrimary := r.Header.Get("From-Primary")
		if fromPrimary == "true" {
			fmt.Println("This Request is from Primary Server")
			addMemo(w, r)
		} else {
			fmt.Println("This Request is from Client")
			forwardMemo(w, r)
		}
	})
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
	router.Use(requestFilter)
	router.HandleFunc("/note", addMemo).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/note/{id}", addMemo).Methods(http.MethodGet, http.MethodDelete, http.MethodPatch)

	fmt.Println("Replica Server is running on port 8081...")
	if err := http.ListenAndServe(":8081", router); err != nil {
		log.Fatal(err)
	}
}

