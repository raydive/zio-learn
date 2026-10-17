package handler

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"zio-learn/internal/playground"
)

func TestExecution(t *testing.T) {
	ts := fakeHTTPServer(t, "cat playground.sc\nprintf '\\n日本語\\n'\nprintf 'stderr\\n' >&2\nexit 7\n", time.Second)
	resp := postTest(t, ts, "/api/run", testRequest)
	defer closeResponse(t, resp)
	body := readBody(t, resp.Body)
	for _, want := range []string{"println(42)", "日本語", "stderr", `"exitCode":7`, `"status":"error"`} {
		if !strings.Contains(string(body), want) {
			t.Errorf("missing %q in %s", want, body)
		}
	}
}

func TestStopAndBusy(t *testing.T) {
	ts := fakeHTTPServer(t, "printf 'ready\\n'\nsleep 30\n", 5*time.Second)
	resp := postTest(t, ts, "/api/run", testRequest)
	defer closeResponse(t, resp)
	reader := bufio.NewReader(resp.Body)
	var id string
	for id == "" {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			t.Fatal(err)
		}
		var event struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(line, &event); err != nil {
			t.Fatal(err)
		}
		id = event.ID
	}
	busy := postTest(t, ts, "/api/run", testRequest)
	closeResponse(t, busy)
	if busy.StatusCode != http.StatusConflict {
		t.Fatalf("expected busy, got %d", busy.StatusCode)
	}
	start := time.Now()
	stop := postTest(t, ts, "/api/stop", map[string]string{"id": id})
	closeResponse(t, stop)
	if stop.StatusCode != http.StatusOK {
		t.Fatal(stop.StatusCode)
	}
	body := readBody(t, reader)
	if !strings.Contains(string(body), `"status":"stopped"`) {
		t.Fatal(string(body))
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("process tree was not stopped promptly")
	}
}

func TestLimits(t *testing.T) {
	for _, tc := range []struct {
		name, script, status string
		timeout              time.Duration
	}{
		{"timeout", "sleep 30\n", "timeout", 50 * time.Millisecond},
		{"output", "yes abcdefghijklmnopqrstuvwxyz\n", "output_limit", 3 * time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ts := fakeHTTPServer(t, tc.script, tc.timeout)
			resp := postTest(t, ts, "/api/run", testRequest)
			defer closeResponse(t, resp)
			body := readBody(t, resp.Body)
			if !strings.Contains(string(body), `"status":"`+tc.status+`"`) {
				t.Fatalf("missing %s in output (%d bytes)", tc.status, len(body))
			}
		})
	}
}

func TestDisconnectCancelsExecution(t *testing.T) {
	ts := fakeHTTPServer(t, "sleep 30\n", time.Minute)
	response := postTest(t, ts, "/api/run", testRequest)
	if response.StatusCode != http.StatusOK {
		t.Fatal(response.StatusCode)
	}
	reader := bufio.NewReader(response.Body)
	if _, err := reader.ReadBytes('\n'); err != nil {
		t.Fatal(err)
	}
	closeResponse(t, response)
	deadline := time.Now().Add(3 * time.Second)
	for {
		next := postTest(t, ts, "/api/run", testRequest)
		status := next.StatusCode
		closeResponse(t, next)
		if status == http.StatusOK {
			return
		}
		if status != http.StatusConflict || time.Now().After(deadline) {
			t.Fatalf("execution was not released after disconnect: %d", status)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestStartFailureReturnsJSON(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	p := playground.New(time.Second)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "http://localhost/api/run", strings.NewReader(`{"code":"object Main extends ZIOAppDefault"}`))
	testHandler(t, p).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusInternalServerError || recorder.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("status=%d headers=%v", recorder.Code, recorder.Header())
	}
}
