package main

import (
	"flag"
	"fmt"
	"log"

	"gopkg.in/yaml.v2"

	"github.com/boltdb/bolt"
)

var dbName string

func parseYAMLMap(y []byte) (map[string]string, error) {
	m := make(map[string]string)

	err := yaml.Unmarshal(y, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func init() {
	flag.StringVar(&dbName, "name", "redirects.db", "name of the BoltDB file to hold redirect information")
}

func main() {
	flag.Parse()

	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	m := map[string]string{"/google": "https://google.com"}

	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("Redirects"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		for key, value := range m {
			if err = b.Put([]byte(key), []byte(value)); err != nil {
				return fmt.Errorf("put: %s", err)
			}
		}

		return nil
	})

}
