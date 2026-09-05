package internal

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGatewayHealthAndProxy(t *testing.T) {
	var sawAuth, sawRequestID string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"ok"}`))
			return
		}
		sawAuth = r.Header.Get("Authorization")
		sawRequestID = r.Header.Get("X-Request-Id")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	t.Cleanup(upstream.Close)

	srv, err := New(Config{
		HTTPAddr:  ":0",
		AuthURL:   upstream.URL,
		UserURL:   upstream.URL,
		MarketURL: upstream.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	gw := httptest.NewServer(srv.Handler())
	t.Cleanup(gw.Close)

	res, err := http.Get(gw.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("health status %d", res.StatusCode)
	}
	_ = res.Body.Close()

	res, err = http.Get(gw.URL + "/ready")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("ready status %d", res.StatusCode)
	}
	_ = res.Body.Close()

	req, err := http.NewRequest(http.MethodGet, gw.URL+"/api/v1/users/me", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-Request-Id", "req-123")
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("proxy status %d body %s", res.StatusCode, body)
	}
	if sawAuth != "Bearer test-token" {
		t.Fatalf("auth not forwarded: %q", sawAuth)
	}
	if sawRequestID != "req-123" {
		t.Fatalf("request id not forwarded: %q", sawRequestID)
	}

	res, err = http.Get(gw.URL + "/api/v1/orders")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown route status %d", res.StatusCode)
	}
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.Error.Code != "NOT_FOUND" {
		t.Fatalf("code %s", payload.Error.Code)
	}
}

func TestGatewayBadGateway(t *testing.T) {
	srv, err := New(Config{
		HTTPAddr:  ":0",
		AuthURL:   "http://127.0.0.1:1",
		UserURL:   "http://127.0.0.1:1",
		MarketURL: "http://127.0.0.1:1",
	})
	if err != nil {
		t.Fatal(err)
	}
	gw := httptest.NewServer(srv.Handler())
	t.Cleanup(gw.Close)

	res, err := http.Get(gw.URL + "/api/v1/auth/login")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadGateway {
		t.Fatalf("status %d", res.StatusCode)
	}
}
