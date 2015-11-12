package main

import (
	"./logging"
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

func TestHttpWorking(t *testing.T) {
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
	expected_hash := "033429aed5b9104f5c8d0a15ed2b9a043ce93a70"

	// First we will read the sample json file
	dat, err := ioutil.ReadFile("test/sample_data.json")
	logging.HandleErr(err)

	// We read the sample json data and create a new DockerInfo struct
	var parsed_data DockerInfo
	err = json.Unmarshal(dat, &parsed_data)
	logging.HandleErr(err)

	// Pass the DockerInfo struct into the GenerateDockerFile function
	GenerateDockerFile(parsed_data, "test")
	defer os.Remove("test/Dockerfile")

	// Generate a sha1 hash of the generated Dockerfile and compare
	f, err := ioutil.ReadFile("test/Dockerfile")
	logging.HandleErr(err)

	generated_hash := sha1.New()
	generated_hash.Write([]byte(f))
	bs := generated_hash.Sum(nil)

	// We would like a hex-encoding string to compare with
	hash_string := hex.EncodeToString(bs[:])

	expect(t, hash_string, expected_hash)
}
