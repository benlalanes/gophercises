package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"
)

var pathsfile string
var useJSON bool

// YAMLHandler returns an HTTP handler that shortens URLs by redirecting
// requests according to the specified YAML configuration.
func YAMLHandler(y []byte, fallback http.Handler) (http.HandlerFunc, error) {
	m := make(map[string]string)
	if err := yaml.Unmarshal(y, &m); err != nil {
		return nil, err
	}

	return mapHandler(m, fallback), nil
}

// JSONHandler returns an HTTP handler that shortens URLs by redirecting
// requests according to the specified JSON configuration.
func JSONHandler(j []byte, fallback http.Handler) (http.HandlerFunc, error) {
	m := make(map[string]string)
	if err := json.Unmarshal(j, &m); err != nil {
		return nil, err
	}

	return mapHandler(m, fallback), nil
}

func mapHandler(redirects map[string]string, fallback http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if to, ok := redirects[r.URL.Path]; ok {
			http.Redirect(w, r, to, http.StatusFound)
		} else {
			fallback.ServeHTTP(w, r)
		}
	})
}

func init() {
	flag.StringVar(&pathsfile, "paths", "", "YAML file containing paths")
	flag.BoolVar(&useJSON, "json", false, "specify to use a JSON file")
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

	var handler http.HandlerFunc

	if useJSON {
		handler, err = JSONHandler(b, fallback)
	} else {
		handler, err = YAMLHandler(b, fallback)
	}

	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe("localhost:8080", handler))
}
