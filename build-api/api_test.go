package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"github.com/go-martini/martini"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestPostBuild(t *testing.T) {
	// This is just an example
	m := martini.Classic()

	m.Get("/foo", func() string {
		return "bar"
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	m.ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Body.String(), `bar`)

}

/* Test Helpers */
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func TestDockerfileCreation(t *testing.T) {
	// The 'expected' hash
	expected_hash := "d34aeddb569f9f15aadd3bb46abf3c39872f659a"

	// First we will read the sample json file
	dat, err := ioutil.ReadFile("test/sample_data.json")
	HandleErr(err)

	var parsed_data DockerInfo

	err = json.Unmarshal(dat, &parsed_data)
	HandleErr(err)

	GenerateDockerFile(parsed_data, "test")
	defer os.Remove("test/Dockerfile")

	// Generate a sha1 hash of the generated Dockerfile and compare
	f, err := ioutil.ReadFile("test/Dockerfile")
	HandleErr(err)

	generated_hash := sha1.New()

	generated_hash.Write([]byte(f))

	bs := generated_hash.Sum(nil)

	hash_string := hex.EncodeToString(bs[:])

	expect(t, hash_string, expected_hash)
}
