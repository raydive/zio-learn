package handler

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestScalaVersionSelection(t *testing.T) {
	ts := fakeHTTPServer(t, "cat playground.sc\n", time.Second)
	for _, path := range []string{"/api/script", "/api/run"} {
		req := runRequest{Code: testRequest.Code, ScalaVersion: "2.13.18"}
		resp := postTest(t, ts, path, req)
		body := readBody(t, resp.Body)
		closeResponse(t, resp)
		if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), `//\u003e using scala 2.13.18`) {
			t.Fatalf("%s did not use selected version: %s", path, body)
		}
		for _, version := range []string{"9.9.9", "3.3.7\n//> using dep bad"} {
			req.ScalaVersion = version
			resp := postTest(t, ts, path, req)
			closeResponse(t, resp)
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("%s accepted invalid version %q", path, version)
			}
		}
	}
}
