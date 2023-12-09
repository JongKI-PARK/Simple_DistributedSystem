package main

import (
    "bytes"
//	"encoding/json"
    "fmt"
    "net/http"
	"time"
	"io/ioutil"
)

func main() {
	var (
		startTime time.Time
		resp      *http.Response
		body      []byte
		err       error
	)

	// GET 요청
	startTime = time.Now()
	getURL := "http://127.0.0.1:8080/note"
	resp, err = http.Get(getURL)
	if err != nil {
		fmt.Println("GET request error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ = ioutil.ReadAll(resp.Body)
	fmt.Println("GET Response Status:", resp.Status)
	fmt.Println("GET Response Body:", string(body))

	elapsed := time.Since(startTime)
	fmt.Println("GET Request took:", elapsed)

	// POST 요청
    postData := []byte(`{"title": "third memo", "body": "third memo body"}`)
    postURL := "http://127.0.0.1:8080/note"
    reqPost, err := http.NewRequest("POST", postURL, bytes.NewBuffer(postData))
    if err != nil {
        fmt.Println("POST request error:", err)
        return
    }
    reqPost.Header.Set("Content-Type", "application/json")
	reqPost.Header.Set("Cache-Control", "no-cache")

    startTime = time.Now()
    resp, err = http.DefaultClient.Do(reqPost)
    if err != nil {
        fmt.Println("POST request error:", err)
        return
    }
    defer resp.Body.Close()

    body, _ = ioutil.ReadAll(resp.Body)
    fmt.Println("POST Response Status:", resp.Status)
    fmt.Println("POST Response Body:", string(body))

    elapsed = time.Since(startTime)
    fmt.Println("POST Request took:", elapsed)

    // DELETE 요청
    deleteURL := "http://127.0.0.1:8080/note/1"
    reqDelete, err := http.NewRequest("DELETE", deleteURL, nil)
    if err != nil {
        fmt.Println("DELETE request error:", err)
        return
    }
	reqDelete.Header.Set("Cache-Control", "no-cache")

	startTime = time.Now()
    resp, err = http.DefaultClient.Do(reqDelete)
    if err != nil {
        fmt.Println("DELETE request error:", err)
        return
    }
    defer resp.Body.Close()

    body, _ = ioutil.ReadAll(resp.Body)
    fmt.Println("DELETE Response Status:", resp.Status)
    fmt.Println("DELETE Response Body:", string(body))

    elapsed = time.Since(startTime)
    fmt.Println("DELETE Request took:", elapsed)

    // PATCH 요청
    patchData := []byte(`{"body": "updated memo body"}`)
    patchURL := "http://127.0.0.1:8080/note/2"
    reqPatch, err := http.NewRequest("PATCH", patchURL, bytes.NewBuffer(patchData))
    if err != nil {
        fmt.Println("PATCH request error:", err)
        return
    }
    reqPatch.Header.Set("Content-Type", "application/json")
	reqPatch.Header.Set("Cache-Control", "no-cache")

    startTime = time.Now()
    resp, err = http.DefaultClient.Do(reqPatch)
    if err != nil {
        fmt.Println("PATCH request error:", err)
        return
    }
    defer resp.Body.Close()

    body, _ = ioutil.ReadAll(resp.Body)
    fmt.Println("PATCH Response Status:", resp.Status)
    fmt.Println("PATCH Response Body:", string(body))

    elapsed = time.Since(startTime)
    fmt.Println("PATCH Request took:", elapsed)
}

