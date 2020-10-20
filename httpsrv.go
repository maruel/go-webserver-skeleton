// Copyright 2019 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Minimalist JSON API server.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"time"

	"github.com/maruel/serve-dir/loghttp"
)

// methodOnly return 405 if the method is not in the allow list.
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
	return methodOnly(f, http.MethodPost)
}

// jsonAPI handles a function, using reflection.
//
// The function passed in must have the form, panics otherwise:
//  foo(out *TypeIn, in *TypeOut) int
//
// The return value must be an HTTP code to return. Normally should be
// http.StatusOK (200).
func jsonAPI(f interface{}) http.HandlerFunc {
	v := reflect.ValueOf(f)
	t := v.Type()
	if t.Kind() != reflect.Func || t.NumIn() != 2 || t.NumOut() != 1 {
		panic("function must have two inputs, zero outputs")
	}
	outT := t.In(0)
	inT := t.In(1)
	if outT.Kind() != reflect.Ptr {
		panic("out must be pointer to struct")
	}
	if outT = outT.Elem(); outT.Kind() != reflect.Struct {
		panic("out must be pointer to struct")
	}
	if inT.Kind() != reflect.Ptr {
		panic("in must be pointer to struct")
	}
	if inT = inT.Elem(); inT.Kind() != reflect.Struct {
		panic("in must be pointer to struct")
	}
	if t.Out(0).Kind() != reflect.Int {
		panic("return value must be integer")
	}
	h := func(w http.ResponseWriter, req *http.Request) {
		data, err := ioutil.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		in := reflect.New(inT)
		if err := json.Unmarshal(data, in.Interface()); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		out := reflect.New(outT)
		args := []reflect.Value{out, in}
		ret := v.Call(args)
		d, _ := json.Marshal(out.Interface())
		w.WriteHeader(int(ret[0].Int()))
		w.Write(d)
	}
	return jsonOnly(h)
}

func mainImpl() error {
	chanSignal := make(chan os.Signal)
	mux := http.NewServeMux()

	// Load our example:
	registerHandlers(mux, chanSignal)

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
		return err
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
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "go-webserver-skeleton: %s\n", err)
		os.Exit(1)
	}
}
