package handlers

import (
	"html/template"
	"io"
	"net/http"
)

// XSS 공격 등 보안 취약점이 생길 수 있기 때문에
// 신뢰할 수 없는 데이터를 writer로 쓸 때에는 항상 html/template 패키지를 사용하는 것이 좋다.
var t = template.Must(template.New("hello").Parse("Hello, {{.}}!"))

func DefaultHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func(r io.ReadCloser) {
				// 바디를 닫으려면 한번은 소비해야 함
				_, _ = io.Copy(io.Discard, r)
				_ = r.Close()
			}(r.Body)

			var b []byte

			switch r.Method {
			case http.MethodGet:
				b = []byte("friend")
			case http.MethodPost:
				var err error
				b, err = io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
			default:
				// "Allow" 헤더가 없기 떄문에 RFC규격을 따르지 않음
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			_ = t.Execute(w, string(b))
		},
	)
}
