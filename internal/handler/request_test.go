package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"zio-learn/internal/playground"
)

func TestRejectInvalidRequest(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	p := playground.New(time.Second)
	handler := testHandler(t, p)
	for _, path := range []string{"/api/script", "/api/run"} {
		for _, tc := range []struct{ name, body, message string }{
			{"empty", `{}`, "コードを入力してください。"},
			{"whitespace", `{"code":"  \n"}`, "コードを入力してください。"},
			{"missing application", `{"code":"println(42)"}`, "ZIOAppDefault を継承する object と、その run を定義してください。"},
			{"multiple applications", `{"code":"object A extends ZIOAppDefault; object B extends ZIOAppDefault"}`, "ZIOAppDefault を継承する object は1つにしてください。"},
			{"unsupported version", `{"code":"object Main extends ZIOAppDefault", "scalaVersion":"9.9.9"}`, "選択できない Scala バージョンです。"},
		} {
			t.Run(path+"/"+tc.name, func(t *testing.T) {
				response := httptest.NewRecorder()
				handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "http://localhost"+path, strings.NewReader(tc.body)))
				if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), tc.message) {
					t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
				}

			})
		}
	}
}
