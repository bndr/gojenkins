package gojenkins

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListUsers(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/securityRealm/api/json", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("depth"); got != "1" {
			t.Fatalf("expected depth=1 query parameter, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"users":{"users":[{"user":{"id":"admin","fullName":"Admin","property":[{"_class":"hudson.tasks.Mailer$UserProperty","address":"admin@example.com"}]}}]}}`)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := server.Client()
	jenk := CreateJenkins(client, server.URL)

	ctx := context.Background()
	users, err := jenk.ListUsers(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].UserName != "admin" {
		t.Fatalf("expected username 'admin', got %s", users[0].UserName)
	}
	if users[0].FullName != "Admin" {
		t.Fatalf("expected full name 'Admin', got %s", users[0].FullName)
	}
	if users[0].Email != "admin@example.com" {
		t.Fatalf("expected email 'admin@example.com', got %s", users[0].Email)
	}
}
