package scrape

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// See api-reverse-engineer.md at the repo root for how these were obtained.
const (
	DefaultGraphQLEndpoint = "https://samizdat-graphql.nytimes.com/graphql/v2"

	userDetailsOperation = "UserDetails"
	userDetailsHash      = "3f462df6ff876e20c737369faf1d3d65725fa54e62ff4039ee13c3e22ecc14f5"

	// nytToken is a static public key the web client sends with every request.
	nytToken = "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAiKjdfob/ixNCvLETwnQ3AalkGSm9NX4gcRbOudrtHmBmIJbWb8Xgu3QH516Edr1qD7A+w+5d0p/WsNCpWDLrqfjTIwMft+jtOQG44l7akD9yi9Gaq/6hS3cuntkY25AYR3WtQPqrtxClX+qQdhMmzlA0sRAXKM8dSbIpsNV9uUOclt3JwB4omwFGj4J+pqzsfYZfB/tlx+BPGjCYGNcZ9O9UvtCpLRLgCJmTugL6V/U581gY8mqp+22aVjbEJik+F0j8xTNSxCOV2PLMpNrRSiDZ8FaKtq8ap/HPey5M7qYZQqclfqsEJMXG/KE3PiaTIbO37caFa80FvzfV8MZw1wIDAQAB"
)

// GraphQLFetcher fetches the friends leaderboard through the NYT Games web
// ("Phoenix") GraphQL API using the UserDetails persisted query.
type GraphQLFetcher struct {
	// Cookie returns the value of the Cookie header to send, verbatim. It is
	// called on every fetch so an updated cookie takes effect without a restart.
	Cookie func(ctx context.Context) (string, error)
	// Endpoint defaults to DefaultGraphQLEndpoint.
	Endpoint string
	// Client defaults to an http.Client with a 10 second timeout.
	Client *http.Client
}

var defaultClient = &http.Client{Timeout: 10 * time.Second}

// DBCookie returns a cookie source that reads the stored cookie from the
// database each time it is called.
func DBCookie(pool *pgxpool.Pool) func(context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		var cookie string
		err := pool.QueryRow(ctx, "SELECT value FROM cookies LIMIT 1").Scan(&cookie)
		return cookie, err
	}
}

type gqlUser struct {
	RegiID  string `json:"regiId"`
	Profile struct {
		GamesUsername string `json:"gamesUsername"`
	} `json:"profile"`
	GameScores struct {
		CrosswordMini *struct {
			SolveTimeSeconds *int32 `json:"solveTimeSeconds"`
		} `json:"crosswordMini"`
	} `json:"gameScores"`
}

type userDetailsResponse struct {
	Data struct {
		User *struct {
			gqlUser
			Friends struct {
				Edges []struct {
					Node gqlUser `json:"node"`
				} `json:"edges"`
			} `json:"friends"`
		} `json:"user"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// Fetch returns the logged-in user and each of their friends for date.
func (f *GraphQLFetcher) Fetch(ctx context.Context, date time.Time) ([]Result, error) {
	cookie, err := f.Cookie(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading cookie: %w", err)
	}

	req, err := f.newRequest(ctx, date, cookie)
	if err != nil {
		return nil, err
	}

	client := f.Client
	if client == nil {
		client = defaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("NYT GraphQL returned %s: %s", resp.Status, snippet)
	}

	var body userDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decoding NYT GraphQL response: %w", err)
	}

	// The API reports partial errors (e.g. "no puzzledata found" for games that
	// did not exist yet on old dates) alongside valid data, so errors are only
	// fatal when there is no user to read.
	user := body.Data.User
	if user == nil {
		msgs := make([]string, len(body.Errors))
		for i, e := range body.Errors {
			msgs[i] = e.Message
		}
		return nil, fmt.Errorf("NYT GraphQL response has no user (errors: %s)", strings.Join(msgs, "; "))
	}

	users := make([]gqlUser, 0, len(user.Friends.Edges)+1)
	users = append(users, user.gqlUser)
	for _, edge := range user.Friends.Edges {
		users = append(users, edge.Node)
	}

	results := make([]Result, 0, len(users))
	for _, u := range users {
		r, err := u.result()
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

func (f *GraphQLFetcher) newRequest(ctx context.Context, date time.Time, cookie string) (*http.Request, error) {
	variables, err := json.Marshal(map[string]string{"printDate": date.UTC().Format(time.DateOnly)})
	if err != nil {
		return nil, err
	}
	extensions, err := json.Marshal(map[string]any{
		"persistedQuery": map[string]any{"sha256Hash": userDetailsHash, "version": 1},
	})
	if err != nil {
		return nil, err
	}

	endpoint := f.Endpoint
	if endpoint == "" {
		endpoint = DefaultGraphQLEndpoint
	}
	query := url.Values{
		"operationName": {userDetailsOperation},
		"variables":     {string(variables)},
		"extensions":    {string(extensions)},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("nyt-app-type", "games-phoenix")
	req.Header.Set("nyt-app-version", "1.0.0")
	req.Header.Set("nyt-token", nytToken)
	req.Header.Set("Cookie", cookie)
	return req, nil
}

func (u gqlUser) result() (Result, error) {
	if u.RegiID == "" {
		return Result{}, errors.New("NYT GraphQL user has no regiId")
	}
	id, err := strconv.ParseInt(u.RegiID, 10, 64)
	if err != nil {
		return Result{}, fmt.Errorf("parsing regiId %q: %w", u.RegiID, err)
	}
	r := Result{NYTUserID: id, Name: u.Profile.GamesUsername}
	// Only a positive time counts; a missing or zero time would otherwise be
	// stored as a permanent 0-second solve.
	if mini := u.GameScores.CrosswordMini; mini != nil && mini.SolveTimeSeconds != nil && *mini.SolveTimeSeconds > 0 {
		r.SolveTimeSeconds = mini.SolveTimeSeconds
	}
	return r, nil
}
