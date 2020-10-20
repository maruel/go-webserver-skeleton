// Copyright 2019 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Minimalist example.
//
// Mean to be shown as "go run ." then run "./example.sh"

package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
)

type logRequest struct {
	Name string `json:"name"`
}

type logResult struct {
	Status int `json:"status"`
}

func apiJSONManual(w http.ResponseWriter, req *http.Request) {
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
	// Call the actual API:
	ret := apiJSONAuto(&out, &in)
	w.WriteHeader(ret)
	d, _ := json.Marshal(&out)
	w.Write(d)
}

func apiJSONAuto(out *logResult, in *logRequest) int {
	switch in.Name {
	case "stdout":
		out.Status = 1
	case "stderr":
		out.Status = 2
	default:
		out.Status = 3
		return http.StatusBadRequest
	}
	return http.StatusOK
}

func registerHandlers(mux *http.ServeMux, c chan os.Signal) {
	quitHandler := func(w http.ResponseWriter, req *http.Request) {
		c <- os.Interrupt
	}
	mux.HandleFunc("/quitquitquit", getOnly(quitHandler))
	mux.HandleFunc("/api/log/manual", jsonOnly(apiJSONManual))
	mux.HandleFunc("/api/log/auto", jsonAPI(apiJSONAuto))
}
