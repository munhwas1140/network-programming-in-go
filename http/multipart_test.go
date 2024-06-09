package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func processMultipart(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("\n %s", string(body))
		t.Log()
		if r.Method != http.MethodPost {
			t.Fatalf("expected method %s; actual status %s",
				http.MethodPost, r.Method)
		}
		// 요청 헤더 출력
		t.Logf("Request Headers\n")
		for name, values := range r.Header {
			t.Logf("%s: %v\n", name, values)
		}
		t.Log("\n")

		reader, err := r.MultipartReader()
		if err != nil {
			t.Fatal(err)
		}

		// 앞에서 바디를 전부 읽어버리면 뒤에서는 읽을 수 없다.
		// 소비되는 개념인 것 같다. 이미지로도 테스트해봐야겠다
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

			if part.FileName() != "" {
				data, err := io.ReadAll(part)
				if err != nil {
					t.Fatal(err)
				}
				t.Logf("File name: %s\n", part.FileName())
				t.Logf("File content: %s\n", string(data))
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
		"./files/goodbye.txt",
	} {

		// MultiPart Section Writer 생성
		filePart, err := w.CreateFormFile(fmt.Sprintf("file%d", i+1),
			filepath.Base(file))
		if err != nil {
			t.Fatal(err)
		}

		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}

		_, err = io.Copy(filePart, f)
		_ = f.Close()
		if err != nil {
			t.Fatal(err)
		}
	}

	err := w.Close()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	ts := httptest.NewServer(http.HandlerFunc(processMultipart(t)))
	defer ts.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		ts.URL, reqBody)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d; actual status %d",
			http.StatusOK, resp.StatusCode)
	}
}
