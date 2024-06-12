package middleware

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPermitPrefix(t *testing.T) {

	handler := http.StripPrefix("/static/",
		PermitPrefix([]string{"sage.svg"}, http.FileServer(http.Dir("../files/"))))

	testCases := []struct {
		path string
		code int
	}{
		{path: "http://test/static/sage.svg", code: http.StatusOK},
		{path: "http://test/static/sage.svg", code: http.StatusBadRequest},
	}

	for _, c := range testCases {
		r := httptest.NewRequest(http.MethodGet, c.path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		b, err := io.ReadAll(w.Body)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(w.Result().StatusCode)
		fmt.Println(string(b))

	}
}
