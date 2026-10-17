package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"zio-learn/internal/assets"
	"zio-learn/internal/playground"
	"zio-learn/internal/transpiler"
)

func fakeHTTPServer(t *testing.T, script string, timeout time.Duration) *httptest.Server {
	t.Helper()
	fakeScalaCLI(t, script)
	p := playground.New(timeout)
	ts := httptest.NewServer(testHandler(t, p))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := p.Shutdown(ctx); err != nil {
			t.Error(err)
		}
		ts.Close()
	})
	return ts
}

func fakeScalaCLI(t *testing.T, script string) string {
	t.Helper()
	dir := t.TempDir()
	cli := filepath.Join(dir, "scala-cli")
	if err := os.WriteFile(cli, []byte("#!/bin/sh\n"+script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return cli
}

func postTest(t *testing.T, ts *httptest.Server, path string, value any) *http.Response {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post(ts.URL+path, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func testConfig(t *testing.T) Config {
	t.Helper()
	data, err := assets.Files().ReadFile("config.json")
	if err != nil {
		t.Fatal(err)
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	config.Examples = make(map[string]string)
	for _, name := range []string{"hello.scala", "tpromise.scala"} {
		code, err := assets.Files().ReadFile("examples/" + name)
		if err != nil {
			t.Fatal(err)
		}
		config.Examples[name] = string(code)
	}
	return config
}

func testHandler(t *testing.T, p *playground.Playground) http.Handler {
	t.Helper()
	config := testConfig(t)
	h, err := New(config, transpiler.New(config.ScalaVersion, config.ScalaVersions, config.Dependencies), p)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func readBody(t *testing.T, reader io.Reader) []byte {
	t.Helper()
	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func closeResponse(t *testing.T, response *http.Response) {
	t.Helper()
	if err := response.Body.Close(); err != nil {
		t.Error(err)
	}
}

var testRequest = runRequest{Code: "object Main extends zio.ZIOAppDefault { def run = zio.ZIO.succeed(println(42)) }"}
