/*
	2023.12.01 ~ 2023.12.12
	Jongki Park

	primary.go : Primary Server code for Distributed System Remote-Write
*/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"bytes"
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
    fmt.Printf("[2023 %s] Primary SERVER [REQUEST]        [METHOD: %s] %s\n", time.Now().Format(time.StampNano), r.Method, message)
}

func getReplicaURLByIndex(index int) (string, error) {
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

    if len(config.Replicas) == 0 || index >= len(config.Replicas) {
        return "", fmt.Errorf("invalid replica index %d", index)
    }

    replicaURL := "http://" + config.Replicas[index] + "/note"

    return replicaURL, nil
}

func syncReplica(r *http.Request, url string, newMemo Memo) (*http.Response, error) {
    if r.Method == http.MethodPost {
		fmt.Printf("[2023 %s] Primary SERVER [UPDATE REPLICA] [METHOD: %s] Request to [%s]\n", time.Now().Format(time.StampNano), r.Method, url)

        postData, err := json.Marshal(map[string]interface{}{
            "title": newMemo.Title,
            "body":  newMemo.Body,
        })
        if err != nil {
            return nil, err
        }

		reqPost, err := http.NewRequest("POST", url, bytes.NewBuffer(postData))
		if err != nil {
			fmt.Println("POST request error:", err)
			return nil, err
		}

		reqPost.Header.Set("From-Primary", "true")
		reqPost.Header.Set("Content-Type", "application/json")
		reqPost.Header.Set("Cache-Control", "no-cache")
		resp, err := http.DefaultClient.Do(reqPost)
		if err != nil {
			fmt.Println("POST request error", err)
			return nil, err
		}
		defer resp.Body.Close()

		fmt.Printf("[2023 %s] Primary SERVER [ACK UPDATE]     [METHOD: %s] Reply from [%s]\n", time.Now().Format(time.StampNano), r.Method, url)

		//body, err := ioutil.ReadAll(resp.Body)
		//fmt.Println("Sync Response from Replica Server POST Response Status:", resp.Status)
		//fmt.Println("Sync Response from Replica Server POST Response Body:", string(body))
		return resp, nil

	} else if r.Method == http.MethodDelete {
		deleteURL := fmt.Sprintf("%s/%d", url, newMemo.ID)
		fmt.Printf("[2023 %s] Primary SERVER [UPDATE REPLICA] [METHOD: %s] Request to [%s]\n", time.Now().Format(time.StampNano), r.Method, deleteURL)

		reqDelete, err := http.NewRequest("DELETE", deleteURL, nil)
		if err != nil {
			fmt.Printf("DELETE request error:", err)
			return nil, err
		}

		reqDelete.Header.Set("From-Primary", "true")
		resp, err := http.DefaultClient.Do(reqDelete)
		if err != nil {
			fmt.Printf("DELETE request error:", err)
			return nil, err
		}
		defer resp.Body.Close()

		fmt.Printf("[2023 %s] Primary SERVER [ACK UPDATE]     [METHOD: %s] Reply from [%s]\n", time.Now().Format(time.StampNano), r.Method, url)

		//body, err := ioutil.ReadAll(resp.Body)
		//fmt.Println("DELETE Response Status:", resp.Status)
		//fmt.Println("DELETE Response Body:", string(body))
		return resp, err

	} else if r.Method == http.MethodPatch {
		patchURL := fmt.Sprintf("%s/%d", url, newMemo.ID)
		fmt.Printf("[2023 %s] Primary SERVER [UPDATE REPLICA] [METHOD: %s] Request to [%s]\n", time.Now().Format(time.StampNano), r.Method, patchURL)

		patchData := make(map[string]interface{})

		// Update the patchData based on newMemo fields
		if newMemo.Title != "" {
			patchData["title"] = newMemo.Title
		}
		if newMemo.Body != "" {
			patchData["body"] = newMemo.Body
		}

		// Marshal patchData to JSON
		jsonData, err := json.Marshal(patchData)
		if err != nil {
			fmt.Println("Error marshaling patchData:", err)
			return nil, err
		}

		reqPatch, err := http.NewRequest("PATCH", patchURL, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("PATCH request error:", err)
			return nil, err
		}

		reqPatch.Header.Set("From-Primary", "true")
		reqPatch.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(reqPatch)
		if err != nil {
			fmt.Println("PATCH request error:", err)
			return nil, err
		}
		defer resp.Body.Close()

		fmt.Printf("[2023 %s] Primary SERVER [ACK REPLICA]    [METHOD: %s] Reply from [%s]\n", time.Now().Format(time.StampNano), r.Method, url)

		//body, err := ioutil.ReadAll(resp.Body)
		//fmt.Println("PATCH Response Status:", resp.Status)
		//fmt.Println("PATCH Response Body:", string(body))
		return resp, err

	} else if r.Method == http.MethodPut {
		putURL := fmt.Sprintf("%s/%d", url, newMemo.ID)
		fmt.Printf("[2023 %s] Primary SERVER [UPDATE REPLICA] [METHOD: %s] Request to [%s]\n", time.Now().Format(time.StampNano), r.Method, putURL)

		putData := make(map[string]interface{})

		// Update the putData based on newMemo fields
		if newMemo.Title != "" {
			putData["title"] = newMemo.Title
		}
		if newMemo.Body != "" {
			putData["body"] = newMemo.Body
		}

		// Marshal putData to JSON
		jsonData, err := json.Marshal(putData)
		if err != nil {
			fmt.Println("Error marshaling putData:", err)
			return nil, err
		}

		reqPut, err := http.NewRequest("PUT", putURL, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("PUT request error:", err)
			return nil, err
		}

		reqPut.Header.Set("From-Primary", "true")
		reqPut.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(reqPut)
		if err != nil {
			fmt.Println("PUT request error:", err)
			return nil, err
		}
		defer resp.Body.Close()

		fmt.Printf("[2023 %s] Primary SERVER [ACK UPDATE]     [METHOD: %s] Reply from [%s]\n", time.Now().Format(time.StampNano), r.Method, url)

		//body, err := ioutil.ReadAll(resp.Body)
		//fmt.Println("PUT Response Status:", resp.Status)
		//fmt.Println("PUT Response Body:", string(body))
		return resp, err

	} else {
		fmt.Printf("Not ready to handle other methods\n")
		return nil, fmt.Errorf("unsupported method: %s", r.Method)
	}
}

func addMemo(w http.ResponseWriter, r *http.Request) {
	replicaURL1, err := getReplicaURLByIndex(1)
	if err != nil {
		return
	}

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

		logRequest(r, "Received new memo with title: ", newMemo.Title)

		response, err := json.Marshal(newMemo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(response)

		message := string(response)
		fmt.Printf("[2023 %s] Primary SERVER [REPLY]          [METHOD: %s] %s\n", time.Now().Format(time.StampNano), r.Method, message)

		resp, err := syncReplica(r, replicaURL1, newMemo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		// do something with the response

	} else if r.Method == http.MethodGet {
		logRequest(r, "Received GET request")
		var message string

		params := mux.Vars(r)
		if idStr, ok := params["id"]; ok {
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}

			memosMu.Lock()
			defer memosMu.Unlock()

			found := false
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

					message = string(response)
					found = true
					break
				}
			}
			if !found {
				message = "No Data"
			}
		} else {
			memosMu.Lock()
			defer memosMu.Unlock()

			if len(memos) == 0 {
				message = "No Data"
			} else {
				response, err := json.Marshal(memos)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(response)

				message = string(response)
			}
		}
		fmt.Printf("[2023 %s] Primary SERVER [REPLY]          [METHOD: %s] %s\n", time.Now().Format(time.StampNano), r.Method, message)

	} else if r.Method == http.MethodDelete {
		logRequest(r, "Received DELETE request")

		params := mux.Vars(r)
		if idStr, ok := params["id"]; ok {
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}

			var newMemo Memo
			newMemo.ID = id

			memosMu.Lock()
			defer memosMu.Unlock()

			for i, memo := range memos {
				if memo.ID == id {
					memos = append(memos[:i], memos[i+1:]...)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"msg": "OK"}`))
					fmt.Printf("[2023 %s] Primary SERVER [REPLY]          [METHOD: %s] %s\n", time.Now().Format(time.StampNano), r.Method, `{"msg": "OK"}`)

					_, err := syncReplica(r, replicaURL1, newMemo)
					if err != nil {
						log.Printf("Failed to sync DELETE request to %s\n", replicaURL1)
					}
					return
				}
			}

			http.Error(w, "Memo not found", http.StatusNotFound)
			fmt.Printf("[2023 %s] Primary SERVER [UPATE REPLICA]  [METHOD: %s] No NEED TO UPDATE REPLICA\n", time.Now().Format(time.StampNano), r.Method)
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

			var newMemo Memo
			newMemo.ID = id
			newMemo.Title = ""
			newMemo.Body = ""

			memosMu.Lock()
			defer memosMu.Unlock()

			for i, memo := range memos {
				if memo.ID == id {
					// Update the memo with the new body if provided in the request
					if newBody, ok := requestBody["body"]; ok {
						memos[i].Body = newBody
						newMemo.Body = newBody
					}

					if newTitle, ok := requestBody["title"]; ok {
						memos[i].Title = newTitle
						newMemo.Title = newTitle
					}

					response, err := json.Marshal(memos[i])
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(response)

					fmt.Printf("[2023 %s] Primary SERVER [REPLY]          [METHOD: %s] %s\n", time.Now().Format(time.StampNano), r.Method, string(response))

					_, err2 := syncReplica(r, replicaURL1, newMemo)
					if err2 != nil {
						log.Printf("Failed to sync PATCH request to %s\n", replicaURL1)
					}
					return
				}
			}

			http.Error(w, "Memo not found", http.StatusNotFound)
			fmt.Printf("[2023 %s] Primary SERVER [UPATE REPLICA]  [METHOD: %s] No NEED TO UPDATE REPLICA\n", time.Now().Format(time.StampNano), r.Method)
			return
		}
		http.Error(w, "Invalid endpoint", http.StatusBadRequest)

	} else if r.Method == http.MethodPut {
		logRequest(r, "Received PUT request")

		params := mux.Vars(r)
		if idStr, ok := params["id"]; ok {
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}

			var newMemo Memo
			err = json.NewDecoder(r.Body).Decode(&newMemo)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			newMemo.ID = id

			memosMu.Lock()
			defer memosMu.Unlock()

			for i, memo := range memos {
				if memo.ID == id {
					memos[i] = newMemo

					response, err := json.Marshal(memos[i])
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(response)

					fmt.Printf("[2023 %s] Primary SERVER [REPLY]          [METHOD: %s] %s\n", time.Now().Format(time.StampNano), r.Method, string(response))

					_, err2 := syncReplica(r, replicaURL1, newMemo)
					if err2 != nil {
						log.Printf("Failed to sync PUT request to %s\n", replicaURL1)
					}
					return
				}
			}

			http.Error(w, "Memo not found", http.StatusNotFound)
			fmt.Printf("[2023 %s] Primary SERVER [UPATE REPLICA]  [METHOD: %s] No NEED TO UPDATE REPLICA\n", time.Now().Format(time.StampNano), r.Method)
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
	router.HandleFunc("/note/{id}", addMemo).Methods(http.MethodGet, http.MethodDelete, http.MethodPatch, http.MethodPut)

	fmt.Println("Primary Server is running on port 8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}

