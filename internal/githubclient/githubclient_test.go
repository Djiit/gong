package githubclient

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/stretchr/testify/assert"
)

const (
	contentTypeHeader   = "Content-Type"
	jsonContentType     = "application/json"
	createdAtTime       = "2023-04-01T12:00:00Z"
	writeResponseErrMsg = "Failed to write response: %v"
)

func TestGetPullRequestState(t *testing.T) {
	// Mock server to simulate GitHub API
	mockServer := createMockServer(t)
	defer mockServer.Close()

	// Create a GitHub client with the mock server
	client := github.NewClient(nil)
	baseURL, _ := url.Parse(mockServer.URL + "/")
	client.BaseURL = baseURL

	// Test different PR states
	testOpenPR(t, client)
	testClosedPR(t, client)
	testMergedPR(t, client)
	testNonExistentPR(t, client)
	testDraftPR(t, client)
}

func createMockServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleMockRequest(t, w, r)
	}))
}

func handleMockRequest(t *testing.T, w http.ResponseWriter, r *http.Request) {
	// Verify the method and path to determine which PR is requested
	if r.Method == "GET" && r.URL.Path == "/repos/testowner/testrepo/pulls/1" {
		serveOpenPR(t, w)
	} else if r.Method == "GET" && r.URL.Path == "/repos/testowner/testrepo/pulls/2" {
		serveClosedPR(t, w)
	} else if r.Method == "GET" && r.URL.Path == "/repos/testowner/testrepo/pulls/3" {
		serveMergedPR(t, w)
	} else if r.Method == "GET" && r.URL.Path == "/repos/testowner/testrepo/pulls/4" {
		serveDraftPR(t, w)
	} else {
		// PR not found
		w.WriteHeader(http.StatusNotFound)
	}
}

func serveOpenPR(t *testing.T, w http.ResponseWriter) {
	w.Header().Set(contentTypeHeader, jsonContentType)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(fmt.Sprintf(`{
		"state": "open",
		"merged": false,
		"created_at": "%s",
		"updated_at": "2023-04-01T13:00:00Z"
	}`, createdAtTime)))
	if err != nil {
		t.Fatalf(writeResponseErrMsg, err)
	}
}

func serveClosedPR(t *testing.T, w http.ResponseWriter) {
	w.Header().Set(contentTypeHeader, jsonContentType)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(fmt.Sprintf(`{
		"state": "closed",
		"merged": false,
		"created_at": "%s",
		"updated_at": "2023-04-02T12:00:00Z"
	}`, createdAtTime)))
	if err != nil {
		t.Fatalf(writeResponseErrMsg, err)
	}
}

func serveMergedPR(t *testing.T, w http.ResponseWriter) {
	w.Header().Set(contentTypeHeader, jsonContentType)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(fmt.Sprintf(`{
		"state": "closed",
		"merged": true,
		"created_at": "%s",
		"updated_at": "2023-04-03T12:00:00Z"
	}`, createdAtTime)))
	if err != nil {
		t.Fatalf(writeResponseErrMsg, err)
	}
}

func serveDraftPR(t *testing.T, w http.ResponseWriter) {
	w.Header().Set(contentTypeHeader, jsonContentType)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(fmt.Sprintf(`{
		"state": "open",
		"merged": false,
		"draft": true,
		"created_at": "%s",
		"updated_at": "2023-04-04T12:00:00Z"
	}`, createdAtTime)))
	if err != nil {
		t.Fatalf(writeResponseErrMsg, err)
	}
}

func testOpenPR(t *testing.T, client *github.Client) {
	t.Run("Open PR", func(t *testing.T) {
		state, err := GetPullRequestState(client, "testowner", "testrepo", "1")
		assert.NoError(t, err)
		assert.NotNil(t, state)
		assert.True(t, state.IsOpen)
		assert.False(t, state.IsClosed)
		assert.False(t, state.IsMerged)

		expectedCreatedAt, _ := time.Parse(time.RFC3339, createdAtTime)
		expectedUpdatedAt, _ := time.Parse(time.RFC3339, "2023-04-01T13:00:00Z")
		assert.Equal(t, expectedCreatedAt, state.CreatedAt)
		assert.Equal(t, expectedUpdatedAt, state.UpdatedAt)
	})
}

func testClosedPR(t *testing.T, client *github.Client) {
	t.Run("Closed PR", func(t *testing.T) {
		state, err := GetPullRequestState(client, "testowner", "testrepo", "2")
		assert.NoError(t, err)
		assert.NotNil(t, state)
		assert.False(t, state.IsOpen)
		assert.True(t, state.IsClosed)
		assert.False(t, state.IsMerged)

		expectedCreatedAt, _ := time.Parse(time.RFC3339, createdAtTime)
		expectedUpdatedAt, _ := time.Parse(time.RFC3339, "2023-04-02T12:00:00Z")
		assert.Equal(t, expectedCreatedAt, state.CreatedAt)
		assert.Equal(t, expectedUpdatedAt, state.UpdatedAt)
	})
}

func testMergedPR(t *testing.T, client *github.Client) {
	t.Run("Merged PR", func(t *testing.T) {
		state, err := GetPullRequestState(client, "testowner", "testrepo", "3")
		assert.NoError(t, err)
		assert.NotNil(t, state)
		assert.False(t, state.IsOpen)
		assert.False(t, state.IsClosed)
		assert.True(t, state.IsMerged)

		expectedCreatedAt, _ := time.Parse(time.RFC3339, createdAtTime)
		expectedUpdatedAt, _ := time.Parse(time.RFC3339, "2023-04-03T12:00:00Z")
		assert.Equal(t, expectedCreatedAt, state.CreatedAt)
		assert.Equal(t, expectedUpdatedAt, state.UpdatedAt)
	})
}

func testNonExistentPR(t *testing.T, client *github.Client) {
	t.Run("Non-existent PR", func(t *testing.T) {
		state, err := GetPullRequestState(client, "testowner", "testrepo", "99")
		assert.Error(t, err)
		assert.Nil(t, state)
	})
}

func testDraftPR(t *testing.T, client *github.Client) {
	t.Run("Draft PR", func(t *testing.T) {
		state, err := GetPullRequestState(client, "testowner", "testrepo", "4")
		assert.NoError(t, err)
		assert.NotNil(t, state)
		assert.True(t, state.IsOpen)
		assert.False(t, state.IsClosed)
		assert.False(t, state.IsMerged)
		assert.True(t, state.IsDraft)

		expectedCreatedAt, _ := time.Parse(time.RFC3339, createdAtTime)
		expectedUpdatedAt, _ := time.Parse(time.RFC3339, "2023-04-04T12:00:00Z")
		assert.Equal(t, expectedCreatedAt, state.CreatedAt)
		assert.Equal(t, expectedUpdatedAt, state.UpdatedAt)
	})
}
