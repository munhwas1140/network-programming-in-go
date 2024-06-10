package main

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func processMultipart(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// body, err := io.ReadAll(r.Body)
		// if err != nil {
		// 	t.Fatal(err)
		// }
		// t.Logf("\n %s", string(body))
		// t.Log()
		if r.Method != http.MethodPost {
			t.Fatalf("expected method %s; actual status %s",
				http.MethodPost, r.Method)
		}

		// 요청 헤더 출력
		// t.Logf("Request Headers\n")
		// for name, values := range r.Header {
		// 	t.Logf("%s: %v\n", name, values)
		// }
		// t.Log("\n")

		reader, err := r.MultipartReader()
		if err != nil {
			t.Fatal(err)
		}

		for {
			part, err := reader.NextPart()
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Fatal(err)
			}

			for name, values := range part.Header {
				t.Logf("%s: %s", name, values)
			}

			if part.FormName() != "" && part.FileName() == "" {
				data, err := io.ReadAll(part)
				if err != nil {
					t.Fatal(err)
				}
				t.Logf("Form field: %s = %s\n", part.FormName(), string(data))
			}

			// 파일 byte가 100 넘어가면 truncate해서 출력
			if part.FileName() != "" {
				data, err := io.ReadAll(part)
				if err != nil {
					t.Fatal(err)
				}
				t.Logf("File name: %s\n", part.FileName())
				if len(data) >= 100 {
					t.Logf("File content(truncate): %s\n", string(data[:100]))
				} else {
					t.Logf("File content: %s\n", string(data))
				}
			}
			t.Log()
		}
		w.WriteHeader(http.StatusOK)
	}
}

func TestMultipartSelfTest(t *testing.T) {
	reqBody := new(bytes.Buffer)
	w := multipart.NewWriter(reqBody)

	for k, v := range map[string]string{
		"date":        time.Now().Format(time.RFC3339),
		"Description": "form values with attached files",
	} {
		err := w.WriteField(k, v)
		if err != nil {
			t.Fatal(err)
		}
	}

	for i, file := range []string{
		"./files/hello.txt",
		"./files/ham.jpg",
	} {

		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		mimeType := mime.TypeByExtension(filepath.Ext(f.Name()))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		filename := filepath.Base(f.Name())
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
				escapeQuotes(fmt.Sprintf("file%d", i)),
				escapeQuotes(filename)))
		h.Set("Content-Type", mimeType)

		part, err := w.CreatePart(h)
		if err != nil {
			t.Fatal(err)
		}

		_, err = io.Copy(part, f)
		if err != nil {
			t.Fatal(err)
		}
	}

	err := w.Close()
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(processMultipart(t)))
	defer ts.Close()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(ts.URL, w.FormDataContentType(), reqBody)
	if err != nil {
		fmt.Println("Request failed:", err)
		return
	}
	defer resp.Body.Close()
}

func escapeQuotes(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}
