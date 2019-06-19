package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/boltdb/bolt"
	"gopkg.in/yaml.v2"
)

var pathsfile string
var useJSON bool
var useBoltDB bool

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

func BoltDBHandler(dbpath string, fallback http.Handler) (http.HandlerFunc, error) {
	m := make(map[string]string)

	db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		return nil, err
	}

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Redirects"))
		if b == nil {
			return errors.New("bucket 'Redirects' does not exist")
		}

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			m[string(k)] = string(v)
		}

		return nil
	})

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
	flag.StringVar(&pathsfile, "paths", "", "file containing paths")
	flag.BoolVar(&useJSON, "json", false, "specify to use a JSON file")
	flag.BoolVar(&useBoltDB, "bolt", false, "specify to use a BoltDB file")
}

func main() {

	flag.Parse()

	if pathsfile == "" {
		log.Fatal("-paths argument is required")
	}

	if useJSON && useBoltDB {
		log.Fatal("only one of useJSON, useBoltDB can be specified")
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
	} else if useBoltDB {
		handler, err = BoltDBHandler("redirects.db", fallback)
	} else {
		handler, err = YAMLHandler(b, fallback)
	}

	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe("localhost:8080", handler))
}
