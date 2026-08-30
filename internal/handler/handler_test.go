package handler

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedirectHandler(t *testing.T) {
	tests := []struct {
		Name           string
		Method         string
		Body           string
		Code           int
		ExpectedMethod string
		ExpectedBody   string
	}{
		{
			Name:           "308 preserves method and body",
			Method:         "POST",
			Body:           `{"foo":"bar"}`,
			Code:           308,
			ExpectedMethod: "POST",
			ExpectedBody:   `{"foo":"bar"}`,
		},
		{
			Name:           "307 preserves method and body",
			Method:         "POST",
			Body:           `{"foo":"bar"}`,
			Code:           307,
			ExpectedMethod: "POST",
			ExpectedBody:   `{"foo":"bar"}`,
		},
		{
			Name:           "303 changes to GET and drops body",
			Method:         "POST",
			Body:           `{"foo":"bar"}`,
			Code:           303,
			ExpectedMethod: "GET",
			ExpectedBody:   "",
		},
		{
			Name:           "302 changes to GET and drops body",
			Method:         "POST",
			Body:           `{"foo":"bar"}`,
			Code:           302,
			ExpectedMethod: "GET",
			ExpectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			var finalReceived struct {
				Method string
				Body   string
			}

			finalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				finalReceived.Method = r.Method
				finalReceived.Body = string(body)
				w.WriteHeader(http.StatusOK)
			}))
			defer finalServer.Close()

			redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, finalServer.URL, tt.Code)
			}))
			defer redirectServer.Close()

			client := &http.Client{}

			req, err := http.NewRequestWithContext(context.Background(), tt.Method, redirectServer.URL, bytes.NewBuffer([]byte(tt.Body)))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if finalReceived.Method != tt.ExpectedMethod {
				t.Errorf("Expected method %s, got %s", tt.ExpectedMethod, finalReceived.Method)
			}
			if finalReceived.Body != tt.ExpectedBody {
				t.Errorf("Expected body %q, got %q", tt.ExpectedBody, finalReceived.Body)
			}
		})
	}
}
