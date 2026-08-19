package argo_upstream

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckerReportsAvailableUpdateAndClassifiesChanges(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/repos/evolution-foundation/evolution-go/commits/main":
			response.Write([]byte(`{"sha":"latest-sha"}`))
		case "/repos/evolution-foundation/evolution-go/tags":
			response.Write([]byte(`[{"name":"0.8.0","commit":{"sha":"latest-sha"}}]`))
		case "/repos/evolution-foundation/evolution-go/compare/base-sha...latest-sha":
			response.Write([]byte(`{"status":"ahead","ahead_by":2,"html_url":"https://github.com/compare","commits":[{"sha":"111111111","html_url":"https://github.com/1","commit":{"message":"sync: 0.8.0"}}],"files":[{"filename":"CHANGELOG.md","patch":"@@ -1 +1,3 @@\n+## 0.8.0\n+- fix: reconnect session\n+- feat: add transport option"}]}`))
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()
	checker := NewChecker(Config{Repository: DefaultRepository, Branch: "main", BaselineSHA: "base-sha", BaselineVersion: "0.7.2", APIURL: server.URL}, server.Client())
	snapshot := checker.Check(t.Context())
	if snapshot.Status != "update_available" || snapshot.BehindBy != 2 || snapshot.LatestVersion != "0.8.0" {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}
	if len(snapshot.Changes) < 2 || snapshot.Changes[0].Category != "fix" || snapshot.Changes[1].Category != "feature" {
		t.Fatalf("unexpected changes: %+v", snapshot.Changes)
	}
}

func TestCheckerPersistsUnavailableStateWithoutLeakingLargeResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		http.Error(response, "rate limited", http.StatusForbidden)
	}))
	defer server.Close()
	checker := NewChecker(Config{Repository: DefaultRepository, Branch: "main", BaselineSHA: "base", APIURL: server.URL}, server.Client())
	snapshot := checker.Check(t.Context())
	if snapshot.Status != "unavailable" || snapshot.Error == "" || len(snapshot.Error) > 500 {
		t.Fatalf("unexpected unavailable snapshot: %+v", snapshot)
	}
}

func TestClassifyUpstreamChanges(t *testing.T) {
	tests := map[string]string{"fix: repair receipt": "fix", "security: update dependency": "security", "feat!: breaking change": "breaking", "chore(deps): bump lib": "maintenance"}
	for title, want := range tests {
		if got := classify(title); got != want {
			t.Fatalf("classify(%q) = %q, want %q", title, got, want)
		}
	}
}
