// +build OMIT

// This is a somewhat cut back version of webfront, available at
// http://github.com/nf/webfront

/*
Copyright 2011 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
webfront is an HTTP server and reverse proxy.

It reads a JSON-formatted rule file like this:

[
	{"Host": "example.com", "Serve": "/var/www"},
	{"Host": "example.org", "Forward": "localhost:8080"}
]

For all requests to the host example.com (or any name ending in
".example.com") it serves files from the /var/www directory.

For requests to example.org, it forwards the request to the HTTP
server listening on localhost port 8080.

Usage of webfront:
  -http=":80": HTTP listen address
  -poll=10s: file poll interval
  -rules="": rule definition file

webfront was written by Andrew Gerrand <adg@golang.org>
*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	httpAddr     = flag.String("http", ":80", "escuchando la direccion HTTP")
	ruleFile     = flag.String("rules", "", "archivo de definicion de reglas")
	pollInterval = flag.Duration("poll", time.Second*10,
		"intervalo de tiempo para la lectura del archivo")
)

func main() {
	flag.Parse()

	s, err := NewServer(*ruleFile, *pollInterval)
	if err != nil {
		log.Fatal(err)
	}

	err = http.ListenAndServe(*httpAddr, s)
	if err != nil {
		log.Fatal(err)
	}
}

// Server implementa http.Handler que actua como un proxy reverso,
// o como un simple servidor de archivo, segun lo determinado por un conjunto de reglas.
type Server struct {
	mu    sync.RWMutex // proteje los siguiente campos
	mtime time.Time    // cuando el archivo de reglas fue modificado por ultima vez
	rules []*Rule
}

// Rule representa una de las reglas en el archivo de configuracion.
type Rule struct {
	Host    string // para comparar la cabecera en la peticion al Host
	Forward string // no puede ser vacio si es reverse proxy
	Serve   string // no puede ser vacio si es un archivo que debe devolver el servidor
}

// Match devuelve true si la regla coincide con la peticion dada.
func (r *Rule) Match(req *http.Request) bool {
	return req.Host == r.Host || strings.HasSuffix(req.Host, "."+r.Host)
}

// Handler devuelve el manejador apropiado para Rule.
func (r *Rule) Handler() http.Handler {
	if h := r.Forward; h != "" {
		return &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = "http"
				req.URL.Host = h
			},
		}
	}
	if d := r.Serve; d != "" {
		return http.FileServer(http.Dir(d))
	}
	return nil
}

// NewServer construye un Server que lee las reglas desde el archivo cada un periodo
// de tiempo espesificado por poll.
func NewServer(file string, poll time.Duration) (*Server, error) {
	s := new(Server)
	if err := s.loadRules(file); err != nil {
		return nil, err
	}
	go s.refreshRules(file, poll)
	return s, nil
}

// ServeHTTP compara las respuesta con Rule, si la encuentra, devuelve la
// respuesta con el manejador de Rule apropiado.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h := s.handler(r); h != nil {
		h.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Not se encontro.", http.StatusNotFound)
}

// handler devuelve el manejador apropiado para la solicitud dada,
// o nil si no lo encontro.
func (s *Server) handler(req *http.Request) http.Handler {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.rules {
		if r.Match(req) {
			return r.Handler()
		}
	}
	return nil
}

// refreshRules lee el archivo de forma periodica y refresca las configuraciones de
// reglas en los Server's si el archivo a sido modificado.
func (s *Server) refreshRules(file string, poll time.Duration) {
	for {
		if err := s.loadRules(file); err != nil {
			log.Println(err)
		}
		time.Sleep(poll)
	}
}

// loadRules comprueba si el archivo ah sido modificado
// y, si lo fue, vuelve a cargar las reglas desde el archivo.
func (s *Server) loadRules(file string) error {
	fi, err := os.Stat(file)
	if err != nil {
		return err
	}
	mtime := fi.ModTime()
	if mtime.Before(s.mtime) && s.rules != nil {
		return nil // no change
	}
	rules, err := parseRules(file)
	if err != nil {
		return fmt.Errorf("parsing %s: %v", file, err)
	}
	s.mu.Lock()
	s.mtime = mtime
	s.rules = rules
	s.mu.Unlock()
	return nil
}

// parseRules lee las definiciones de reglas desde el archivo y devuelve el
// Rule que resulta de ello.
func parseRules(file string) ([]*Rule, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var rules []*Rule
	err = json.NewDecoder(f).Decode(&rules)
	if err != nil {
		return nil, err
	}
	return rules, nil
}
