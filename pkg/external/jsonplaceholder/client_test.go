package jsonplaceholder

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPostSuccess(t *testing.T) {
	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"userId":1,"id":2,"title":"hello","body":"world"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	post, err := client.GetPost(context.Background(), "2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if requestedPath != "/posts/2" {
		t.Fatalf("unexpected path: %s", requestedPath)
	}
	if post.ID != 2 || post.UserID != 1 || post.Title != "hello" || post.Body != "world" {
		t.Fatalf("unexpected post: %#v", post)
	}
}

func TestGetPostMissingID(t *testing.T) {
	client := NewClient(defaultBaseURL, nil)
	if _, err := client.GetPost(context.Background(), " "); err != ErrMissingPostID {
		t.Fatalf("expected ErrMissingPostID, got %v", err)
	}
}

func TestGetPostUnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	if _, err := client.GetPost(context.Background(), "1"); err == nil {
		t.Fatalf("expected error when status not 200")
	}
}

func TestGetPostDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not-json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	if _, err := client.GetPost(context.Background(), "3"); err == nil {
		t.Fatalf("expected decode error")
	}
}
