package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"zio-learn/internal/playground"
)

func TestEmbeddedResources(t *testing.T) {
	p := playground.New(time.Second)
	handler := testHandler(t, p)
	for _, tc := range []struct{ path, content string }{
		{"/", "Scala バージョン"},
		{"/app.js", "scalaVersion"},
		{"/style.css", ".toolbar"},
		{"/api/config", "3.10.0"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "http://localhost"+tc.path, nil))
			if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), tc.content) {
				t.Fatalf("embedded resource %s: status=%d body=%s", tc.path, response.Code, response.Body.String())
			}
		})
	}
}

func TestOriginAndInput(t *testing.T) {
	ts := fakeHTTPServer(t, "exit 0\n", time.Second)
	for _, tc := range []struct {
		host, origin, body string
		status             int
	}{
		{"localhost", "http://foreign.example", `{}`, 403},
		{"foreign.example", "", `{}`, 403},
		{"localhost", "", strings.Repeat("x", 256*1024+1), 400},
		{"localhost", "", `{} {}`, 400},
	} {
		req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/run", strings.NewReader(tc.body))
		if err != nil {
			t.Fatal(err)
		}
		req.Host = tc.host
		req.Header.Set("Origin", tc.origin)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		closeResponse(t, resp)
		if resp.StatusCode != tc.status {
			t.Errorf("got %d, want %d", resp.StatusCode, tc.status)
		}
	}
}
