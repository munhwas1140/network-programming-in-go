package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// 클라이언트는 서버로부터 먼저 응답 상태 코드를 받은 후에
// 응답 바디를 받는다.
//
// 만약 응답 바디르 먼저 쓰면 Go는 응답 상태 코드를 200이라고
// 생각하고 실제로 응답 바디를 보내기 전에 클라이언트에게 먼저 보낸다.
func TestHandlerWriteHeader(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Bad request"))
		w.WriteHeader(http.StatusBadRequest)
	}

	r := httptest.NewRequest(http.MethodGet, "http://test", nil)
	w := httptest.NewRecorder()
	handler(w, r)
	t.Logf("Response status: %q", w.Result().Status)

	handler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Bad request"))
	}

	r = httptest.NewRequest(http.MethodGet, "http://test", nil)
	w = httptest.NewRecorder()
	handler(w, r)
	t.Logf("Response status: %q", w.Result().Status)
}
