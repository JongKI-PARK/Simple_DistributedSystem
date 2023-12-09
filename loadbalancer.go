package main

import (
	"fmt"
	//"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func loadBalancerHandler(targetUrls []*url.URL) func(http.ResponseWriter, *http.Request) {
	nextIndex := 0
	proxyServers := make([]*httputil.ReverseProxy, len(targetUrls))

	for i, targetUrl := range targetUrls {
		proxyServers[i] = httputil.NewSingleHostReverseProxy(targetUrl)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[2023 %s] Load Balancer [FORWARD REQUEST] [METHOD: %s] to [%s]\n", time.Now().Format(time.StampNano), r.Method, targetUrls[nextIndex])

		proxyServers[nextIndex].ServeHTTP(w, r)
		nextIndex = (nextIndex + 1) % len(targetUrls)

		fmt.Printf("[2023 %s] Load Balancer [FORWARD REPLY]   [METHOD: %s] from [%s]\n", time.Now().Format(time.StampNano), r.Method, targetUrls[nextIndex])
	}
}

func main() {
	clientUrls := []*url.URL{
		{
			Scheme: "http",
			Host:   "localhost:8080",
		},
		{
			Scheme: "http",
			Host:   "localhost:8081",
		},
	}

	handler := loadBalancerHandler(clientUrls)

	http.HandleFunc("/", handler)

	fmt.Println("Load Balancer Server is running on port 5000...")
	if err := http.ListenAndServe(":5000", nil); err != nil {
		log.Fatal(err)
	}
}

