package scrape

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

// userDetailsFixture mirrors a real response shape: the logged-in user plus
// friends, one without a Mini time, and a partial error for a game that did not
// exist yet on the requested date.
const userDetailsFixture = `{
  "data": {
    "user": {
      "regiId": "100",
      "profile": {"gamesUsername": "Me", "username": null},
      "setting": null,
      "gameScores": {"connections": null, "crosswordMini": {"solveTimeSeconds": 42}, "spellingBee": null, "wordle": null},
      "friends": {"edges": [
        {"node": {"regiId": "200", "profile": {"gamesUsername": "Fast"}, "gameScores": {"crosswordMini": {"solveTimeSeconds": 17}}}, "relationshipStatus": "FRIEND"},
        {"node": {"regiId": "300", "profile": {"gamesUsername": "Absent"}, "gameScores": {"crosswordMini": null}}, "relationshipStatus": "FRIEND"},
        {"node": {"regiId": "400", "profile": {"gamesUsername": "NoTime"}, "gameScores": {"crosswordMini": {}}}, "relationshipStatus": "FRIEND"},
        {"node": {"regiId": "500", "profile": {"gamesUsername": "Zero"}, "gameScores": {"crosswordMini": {"solveTimeSeconds": 0}}}, "relationshipStatus": "FRIEND"}
      ]}
    }
  },
  "errors": [{"message": "GetGameStats.getGameScoresWorker:FetchPuzzleIDByPrintDate : Game: [connections], Error: [no puzzledata found for print date]"}]
}`

func staticCookie(c string) func(context.Context) (string, error) {
	return func(context.Context) (string, error) { return c, nil }
}

func int32p(v int32) *int32 { return &v }

func TestGraphQLFetcher(t *testing.T) {
	const cookie = "NYT-S=abc; nyt-a=xyz"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("operationName"); got != "UserDetails" {
			t.Errorf("operationName = %q", got)
		}
		var vars map[string]string
		if err := json.Unmarshal([]byte(q.Get("variables")), &vars); err != nil || vars["printDate"] != "2023-01-01" {
			t.Errorf("variables = %q (%v)", q.Get("variables"), err)
		}
		if !strings.Contains(q.Get("extensions"), userDetailsHash) {
			t.Errorf("extensions = %q, want persisted query hash", q.Get("extensions"))
		}
		for header, want := range map[string]string{
			"Cookie":          cookie,
			"nyt-app-type":    "games-phoenix",
			"nyt-app-version": "1.0.0",
			"nyt-token":       nytToken,
		} {
			if got := r.Header.Get(header); got != want {
				t.Errorf("header %s = %q, want %q", header, got, want)
			}
		}
		w.Write([]byte(userDetailsFixture))
	}))
	defer srv.Close()

	f := &GraphQLFetcher{Cookie: staticCookie(cookie), Endpoint: srv.URL, Client: srv.Client()}
	got, err := f.Fetch(context.Background(), time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	want := []Result{
		{NYTUserID: 100, Name: "Me", SolveTimeSeconds: int32p(42)},
		{NYTUserID: 200, Name: "Fast", SolveTimeSeconds: int32p(17)},
		{NYTUserID: 300, Name: "Absent"},
		{NYTUserID: 400, Name: "NoTime"},
		{NYTUserID: 500, Name: "Zero"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Fetch =\n%+v\nwant\n%+v", got, want)
	}
}

func TestGraphQLFetcherErrors(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		wantErr string
	}{
		{"http error", http.StatusNotFound, `not found`, "404"},
		{"not json", http.StatusOK, `<html>`, "decoding"},
		{"no user", http.StatusOK, `{"data": {"user": null}, "errors": [{"message": "unauthorized"}]}`, "unauthorized"},
		{"bad regiId", http.StatusOK, `{"data": {"user": {"regiId": "abc", "friends": {"edges": []}}}}`, "regiId"},
		{"missing regiId", http.StatusOK, `{"data": {"user": {"friends": {"edges": []}}}}`, "regiId"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			f := &GraphQLFetcher{Cookie: staticCookie("c"), Endpoint: srv.URL, Client: srv.Client()}
			_, err := f.Fetch(context.Background(), time.Now())
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Fetch error = %v, want one containing %q", err, tt.wantErr)
			}
		})
	}
}
