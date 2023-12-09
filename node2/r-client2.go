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
	getURL := "http://127.0.0.1:8081/note"
	//fmt.Printf("Request to Address : %s\n", getURL)
	fmt.Printf("[%s] CLIENT [REQUEST] [GET] %s\n", time.Now().Format(time.StampNano), getURL)

	startTime = time.Now()
	resp, err = http.Get(getURL)
	if err != nil {
		fmt.Println("GET request error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ = ioutil.ReadAll(resp.Body)
	elapsed := time.Since(startTime)

	//fmt.Println("GET Response Status:", resp.Status)
	//fmt.Println("GET Response Body:", string(body))
	fmt.Printf("[%s] CLIENT [REPLY] [GET] %s\n", time.Now().Format(time.StampNano), string(body))
	//fmt.Println("GET Response Status:", resp.Status)
	//fmt.Println("GET Response Body:", string(body))
	fmt.Println("GET Request took:", elapsed)


	// POST 요청
    postData := []byte(`{"title": "memo title", "body": "memo body"}`)
    postURL := "http://127.0.0.1:8081/note"
	//fmt.Printf("Request to Address : %s\n", postURL)

    reqPost, err := http.NewRequest("POST", postURL, bytes.NewBuffer(postData))
    if err != nil {
        fmt.Println("POST request error:", err)
        return
    }
    reqPost.Header.Set("Content-Type", "application/json")
	reqPost.Header.Set("Cache-Control", "no-cache")

	fmt.Printf("[%s] CLIENT [REQUEST] [POST] %s {\"title\": \"memo title\", \"body\": \"memo body\"}\n", time.Now().Format(time.StampNano), postURL)

    startTime = time.Now()
    resp, err = http.DefaultClient.Do(reqPost)
    if err != nil {
        fmt.Println("POST request error:", err)
        return
    }
    defer resp.Body.Close()

    body, _ = ioutil.ReadAll(resp.Body)
    //fmt.Println("POST Response Status:", resp.Status)
    //fmt.Println("POST Response Body:", string(body))

    elapsed = time.Since(startTime)
	fmt.Printf("[%s] CLIENT [REPLY] [POST] %s %s\n", time.Now().Format(time.StampNano), postURL, string(body))
	fmt.Println("POST Request took:", elapsed)


    // DELETE 요청
    deleteURL := "http://127.0.0.1:8081/note/1"
	//fmt.Printf("Request to Address : %s\n", deleteURL)

    reqDelete, err := http.NewRequest("DELETE", deleteURL, nil)
    if err != nil {
        fmt.Println("DELETE request error:", err)
        return
    }
	reqDelete.Header.Set("Cache-Control", "no-cache")

	fmt.Printf("[%s] CLIENT [REQUEST] [DELETE] %s\n", time.Now().Format(time.StampNano), deleteURL)
	startTime = time.Now()
    resp, err = http.DefaultClient.Do(reqDelete)
    if err != nil {
        fmt.Println("DELETE request error:", err)
        return
    }
    defer resp.Body.Close()

    body, _ = ioutil.ReadAll(resp.Body)
    //fmt.Println("DELETE Response Status:", resp.Status)
    //fmt.Println("DELETE Response Body:", string(body))

    elapsed = time.Since(startTime)
	fmt.Printf("[%s] CLIENT [REPLY] [DELETE] %s\n", time.Now().Format(time.StampNano), string(body))
    fmt.Println("DELETE Request took:", elapsed)


    // PATCH 요청
    patchData := []byte(`{"body": "updated memo body"}`)
    patchURL := "http://127.0.0.1:8081/note/2"
	//fmt.Printf("Request to Address : %s\n", patchURL)

    reqPatch, err := http.NewRequest("PATCH", patchURL, bytes.NewBuffer(patchData))
    if err != nil {
        fmt.Println("PATCH request error:", err)
        return
    }
    reqPatch.Header.Set("Content-Type", "application/json")
	reqPatch.Header.Set("Cache-Control", "no-cache")

	fmt.Printf("[%s] CLIENT [REQUEST] [PATCH] %s %s\n", time.Now().Format(time.StampNano), patchURL, patchData)

    startTime = time.Now()
    resp, err = http.DefaultClient.Do(reqPatch)
    if err != nil {
        fmt.Println("PATCH request error:", err)
        return
    }
    defer resp.Body.Close()

    body, _ = ioutil.ReadAll(resp.Body)
    //fmt.Println("PATCH Response Status:", resp.Status)
    //fmt.Println("PATCH Response Body:", string(body))

    elapsed = time.Since(startTime)
	fmt.Printf("[%s] CLIENT [REPLY] [PATCH] %s\n", time.Now().Format(time.StampNano), string(body))

    fmt.Println("PATCH Request took:", elapsed)


	// PUT 요청
	putData := []byte(`{"title": "updated memo title", "body": "updated memo body"}`)
	putURL := "http://127.0.0.1:8081/note/2"
	//fmt.Printf("Request to Address : %s\n", putURL)

	reqPut, err := http.NewRequest("PUT", putURL, bytes.NewBuffer(putData))
	if err != nil {
		fmt.Println("PUT request error:", err)
		return
	}
	reqPut.Header.Set("Content-Type", "application/json")
	reqPut.Header.Set("Cache-Control", "no-cache")

	fmt.Printf("[%s] CLIENT [REQUEST] [PUT] %s %s\n", time.Now().Format(time.StampNano), putURL, putData)

	startTime = time.Now()
	resp, err = http.DefaultClient.Do(reqPut)
	if err != nil {
		fmt.Println("PUT request error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ = ioutil.ReadAll(resp.Body)
	//fmt.Println("PUT Response Status:", resp.Status)
	//fmt.Println("PUT Response Body:", string(body))

	elapsed = time.Since(startTime)
	fmt.Printf("[%s] CLIENT [REPLY] [PUT] %s\n", time.Now().Format(time.StampNano), string(body))

	fmt.Println("PUT Request took:", elapsed)
}

