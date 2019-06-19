package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"
)

var pathsfile string

// YAMLHandler returns an HTTP handler that shortens URLs by redirecting
// requests according to the specified YAML configuration.
func YAMLHandler(y []byte, fallback http.Handler) (http.HandlerFunc, error) {
	m := make(map[string]string)
	if err := yaml.Unmarshal(y, &m); err != nil {
		return nil, err
	}

	f := func(w http.ResponseWriter, r *http.Request) {
		if to, ok := m[r.URL.Path]; ok {
			http.Redirect(w, r, to, http.StatusFound)
		} else {
			fallback.ServeHTTP(w, r)
		}
	}

	return http.HandlerFunc(f), nil
}

func init() {
	flag.StringVar(&pathsfile, "paths", "", "YAML file containing paths")
}

func main() {

	flag.Parse()

	if pathsfile == "" {
		log.Fatal("-paths argument is required")
	}

	f, err := os.Open(pathsfile)
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	fallback := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "No redirect found for specified path.")
	})

	handler, err := YAMLHandler(b, fallback)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe("localhost:8080", handler))
}
