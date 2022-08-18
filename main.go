package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	args := os.Args

	if len(args) != 2 {
		fmt.Println("Usage: passthough url")
		os.Exit(1)
	}

	url := args[1]

	client := http.Client{}

	handler := func(w http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		url := fmt.Sprintf("%s%s", url, req.URL.Path)

		fmt.Println(url)

		proxyReq, err := http.NewRequestWithContext(context.Background(), req.Method, url, bytes.NewReader(body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		proxyReq.Header = make(http.Header)
		for h, val := range req.Header {
			proxyReq.Header[h] = val
		}

		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}

		w.WriteHeader(resp.StatusCode)

		_, err = io.Copy(w, resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		resp.Body.Close()
	}

	http.HandleFunc("/", handler)

	panic(http.ListenAndServe(":8881", nil))
}
