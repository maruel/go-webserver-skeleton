// Copyright 2019 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/maruel/serve-dir/loghttp"
)

func methodOnly(f http.HandlerFunc, methods ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		for _, m := range methods {
			if req.Method == m {
				f(w, req)
				return
			}
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// postOnly enforces POST HTTP method.
func postOnly(f http.HandlerFunc) http.HandlerFunc {
	return methodOnly(f, http.MethodPost)
}

// getOnly enforces GET HTTP method.
func getOnly(f http.HandlerFunc) http.HandlerFunc {
	return methodOnly(f, http.MethodGet)
}

// jsonOnly enforces JSON POST RPCs.
func jsonOnly(f http.HandlerFunc) http.HandlerFunc {
	return postOnly(func(w http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "Content type", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		f(w, req)
	})
	return methodOnly(f, http.MethodGet)
}

// Example

type logRequest struct {
	Name string `json:"name"`
}

type logResult struct {
	Status int `json:"status"`
}

func apiJSON(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	in := logRequest{}
	if err := json.Unmarshal(data, &in); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	out := logResult{}
	switch in.Name {
	case "stdout":
		out.Status = 1
	case "stderr":
		out.Status = 2
	default:
		out.Status = 3
	}
	d, _ := json.Marshal(&out)
	w.Write(d)
}

func main() {
	chanSignal := make(chan os.Signal)
	quitHandler := func(w http.ResponseWriter, req *http.Request) {
		chanSignal <- os.Interrupt
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/quitquitquit", getOnly(quitHandler))
	mux.HandleFunc("/api/log", jsonOnly(apiJSON))
	s := &http.Server{
		// loghttp is optional but really helpful:
		Handler:        &loghttp.Handler{Handler: mux},
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	port := ":8081"
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", port, err)
	}
	go func() {
		<-chanSignal
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		s.Shutdown(ctx)
	}()
	signal.Notify(chanSignal, os.Interrupt)

	log.Printf("Serving on %s", port)
	s.Serve(ln)
	log.Printf("Quitting")
}
