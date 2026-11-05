package server_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Yandex-Practicum/go1fl-sprint6-final/internal/server"
)

// Результаты загрузок сохраняются в отдельном временном каталоге.
func startServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	index, err := os.ReadFile(filepath.Join("..", "..", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile("index.html", index, 0600); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewUnstartedServer(nil)
	srv.Config = server.New(log.New(io.Discard, "", 0)).HTTPServer
	srv.Start()
	t.Cleanup(srv.Close)
	return srv, dir
}

func multipartBody(t *testing.T, field, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile(field, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return &body, w.FormDataContentType()
}

func send(t *testing.T, client *http.Client, method, endpoint, contentType string, body io.Reader) (int, http.Header, string) {
	t.Helper()
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, resp.Header, string(data)
}

func outputFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		if entry.Name() != "index.html" {
			names = append(names, entry.Name())
		}
	}
	return names
}

func TestHTTPIndex(t *testing.T) {
	srv, _ := startServer(t)
	status, headers, body := send(t, srv.Client(), "GET", srv.URL+"/", "", nil)
	if status != http.StatusOK || !strings.Contains(body, "<form") || !strings.HasPrefix(headers.Get("Content-Type"), "text/html") {
		t.Fatalf("GET /: status=%d, Content-Type=%q, body=%q", status, headers.Get("Content-Type"), body)
	}
	status, _, body = send(t, srv.Client(), "HEAD", srv.URL+"/", "", nil)
	if status != http.StatusOK || body != "" {
		t.Fatalf("HEAD /: status=%d, body=%q", status, body)
	}
}

func TestHTTPMethodAllowHeaders(t *testing.T) {
	for _, tt := range []struct{ method, path, allow string }{
		{"POST", "/", "GET, HEAD"},
		{"GET", "/upload", "POST"},
		{"HEAD", "/upload", "POST"},
	} {
		t.Run(tt.method+tt.path, func(t *testing.T) {
			srv, _ := startServer(t)
			status, headers, _ := send(t, srv.Client(), tt.method, srv.URL+tt.path, "", nil)
			if status != http.StatusMethodNotAllowed || headers.Get("Allow") != tt.allow {
				t.Fatalf("status=%d Allow=%q; want 405, %q", status, headers.Get("Allow"), tt.allow)
			}
		})
	}
}

func TestHTTPUploadSupportedInputs(t *testing.T) {
	tests := []struct{ name, filename, input, want string }{
		{"text", "input.txt", "Привет", ".--. .-. .. .-- . -"},
		{"morse", "input.txt", ".--. .-. .. .-- . -", "ПРИВЕТ"},
		{"without_extension", "input", "А", ".-"},
		{"unicode_filename", "пример.txt", "А", ".-"},
		{"path_in_filename", "../../index.html", "А", ".-"},
		{"binary_extension_with_text", "input.bin", "А", ".-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, dir := startServer(t)
			body, contentType := multipartBody(t, "myFile", tt.filename, []byte(tt.input))
			status, headers, got := send(t, srv.Client(), "POST", srv.URL+"/upload", contentType, body)
			if status != http.StatusOK || got != tt.want {
				t.Fatalf("POST /upload: status=%d, body=%q; want 200, %q", status, got, tt.want)
			}
			if got := headers.Get("Content-Type"); got != "text/plain; charset=utf-8" {
				t.Fatalf("Content-Type=%q; want text/plain; charset=utf-8", got)
			}
			files := outputFiles(t, dir)
			if len(files) != 1 || !strings.HasSuffix(files[0], filepath.Ext(tt.filename)) {
				t.Fatalf("unexpected output files: %q", files)
			}
			saved, err := os.ReadFile(filepath.Join(dir, files[0]))
			if err != nil || string(saved) != tt.want {
				t.Fatalf("saved result=%q, err=%v; want %q", saved, err, tt.want)
			}
		})
	}
}

func TestHTTPUploadRejectedInputs(t *testing.T) {
	tests := []struct {
		name, method, contentType, raw, field, input string
		multipart                                    bool
		want                                         int
	}{
		{name: "get", method: "GET", want: 405},
		{name: "put", method: "PUT", want: 405},
		{name: "delete", method: "DELETE", want: 405},
		{name: "options", method: "OPTIONS", want: 405},
		{name: "missing_body", method: "POST", want: 500},
		{name: "plain_text", method: "POST", contentType: "text/plain", raw: "А", want: 500},
		{name: "missing_boundary", method: "POST", contentType: "multipart/form-data", raw: "broken", want: 500},
		{name: "truncated_multipart", method: "POST", contentType: "multipart/form-data; boundary=x", raw: "--x\r\n", want: 500},
		{name: "wrong_field", method: "POST", multipart: true, field: "file", input: "А", want: 500},
		{name: "empty_file", method: "POST", multipart: true, field: "myFile", want: 500},
		{name: "whitespace_file", method: "POST", multipart: true, field: "myFile", input: " \r\n\t", want: 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, dir := startServer(t)
			var body io.Reader = strings.NewReader(tt.raw)
			contentType := tt.contentType
			if tt.multipart {
				body, contentType = multipartBody(t, tt.field, "input.txt", []byte(tt.input))
			}
			status, _, _ := send(t, srv.Client(), tt.method, srv.URL+"/upload", contentType, body)
			if status != tt.want {
				t.Fatalf("status=%d; want %d", status, tt.want)
			}
			if files := outputFiles(t, dir); len(files) != 0 {
				t.Fatalf("rejected request created files: %q", files)
			}
		})
	}
}

func TestHTTPMissingIndex(t *testing.T) {
	srv, _ := startServer(t)
	if err := os.Remove("index.html"); err != nil {
		t.Fatal(err)
	}
	status, _, _ := send(t, srv.Client(), "GET", srv.URL+"/", "", nil)
	if status != http.StatusInternalServerError {
		t.Fatalf("status=%d; want 500", status)
	}
}

func TestHTTPConcurrentUploads(t *testing.T) {
	srv, dir := startServer(t)
	const count = 20
	errors := make(chan error, count)
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		body, contentType := multipartBody(t, "myFile", "input.txt", []byte("А"))
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := srv.Client().Post(srv.URL+"/upload", contentType, body)
			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()
			data, err := io.ReadAll(resp.Body)
			if err != nil || resp.StatusCode != 200 || string(data) != ".-" {
				errors <- fmt.Errorf("status=%d body=%q err=%v", resp.StatusCode, data, err)
			}
		}()
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
	if files := outputFiles(t, dir); len(files) != count {
		t.Fatalf("saved %d results for %d uploads", len(files), count)
	}
}

func TestHTTPIncompleteUploadReadTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("waits for the configured 5-second read timeout")
	}
	srv, dir := startServer(t)
	conn, err := net.DialTimeout("tcp", srv.Listener.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(8 * time.Second)); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	_, err = io.WriteString(conn, "POST /upload HTTP/1.1\r\nHost: test\r\nContent-Type: multipart/form-data; boundary=x\r\nContent-Length: 100\r\nConnection: close\r\n\r\n--x\r\n")
	if err != nil {
		t.Fatal(err)
	}
	response, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("incomplete request did not terminate within 8 seconds: %v", err)
	}
	elapsed := time.Since(started)
	if !bytes.HasPrefix(response, []byte("HTTP/1.1 500 ")) || elapsed < 4*time.Second || elapsed > 7*time.Second {
		t.Fatalf("incomplete upload: elapsed=%s, response=%q; want 500 near 5 seconds", elapsed, response)
	}
	if files := outputFiles(t, dir); len(files) != 0 {
		t.Fatalf("incomplete upload created files: %q", files)
	}
}

func TestHTTPUploadAndRouting(t *testing.T) {
	tests := []struct {
		name, input, want string
		status            int
	}{
		{"word_boundary", "А Б", ".-   -...", 200},
		{"multiline_morse", "...\n---\n...", "СОС", 200},
		{"unsupported_text", "🙂", "", 500},
		{"unknown_morse", ".......", "", 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, dir := startServer(t)
			body, contentType := multipartBody(t, "myFile", "input.txt", []byte(tt.input))
			status, _, got := send(t, srv.Client(), "POST", srv.URL+"/upload", contentType, body)
			if status != tt.status || (tt.status == 200 && got != tt.want) {
				t.Fatalf("input=%q: status=%d body=%q; want status=%d body=%q", tt.input, status, got, tt.status, tt.want)
			}
			if tt.status != http.StatusOK && len(outputFiles(t, dir)) != 0 {
				t.Fatal("invalid input created an output file")
			}
		})
	}
	for _, path := range []string{"/missing", "/upload/"} {
		t.Run("unknown_route"+path, func(t *testing.T) {
			srv, _ := startServer(t)
			status, _, body := send(t, srv.Client(), "GET", srv.URL+path, "", nil)
			if status != http.StatusNotFound {
				t.Fatalf("GET %s: status=%d, HTML=%t; want 404", path, status, strings.Contains(body, "<form"))
			}
		})
	}
	t.Run("form_uses_current_origin", func(t *testing.T) {
		srv, _ := startServer(t)
		_, _, html := send(t, srv.Client(), "GET", srv.URL+"/", "", nil)
		match := regexp.MustCompile(`action="([^"]+)"`).FindStringSubmatch(html)
		if len(match) != 2 {
			t.Fatal("missing form action")
		}
		base, _ := url.Parse(srv.URL)
		action, err := url.Parse(match[1])
		if err != nil {
			t.Fatal(err)
		}
		target := base.ResolveReference(action)
		if target.Host != base.Host || target.Scheme != base.Scheme || target.Path != "/upload" {
			t.Fatalf("page at %s submits to %s", base, target)
		}
	})
}
