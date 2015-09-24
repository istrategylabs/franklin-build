package main

import (
	"github.com/go-martini/martini"
	"net/http"
	"net/http/httptest"
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
