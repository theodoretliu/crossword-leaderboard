package scrape

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"
)

type saveCall struct {
	users  []Result
	solves []solve
}

func recordingBatcher(t *testing.T, fail *bool) (*Batcher, *[]saveCall) {
	t.Helper()
	var calls []saveCall
	b := newBatcher(func(_ context.Context, users []Result, solves []solve) error {
		slices.SortFunc(solves, func(a, b solve) int {
			if c := a.Date.Compare(b.Date); c != 0 {
				return c
			}
			return int(a.NYTUserID - b.NYTUserID)
		})
		calls = append(calls, saveCall{users, solves})
		if fail != nil && *fail {
			return errors.New("db down")
		}
		return nil
	})
	return b, &calls
}

var (
	day1 = time.Date(2025, 5, 19, 0, 0, 0, 0, time.UTC)
	day2 = time.Date(2025, 5, 20, 0, 0, 0, 0, time.UTC)
)

func TestBatcherSkipsWhatIsStored(t *testing.T) {
	b, calls := recordingBatcher(t, nil)
	b.names[1] = "Alice"
	b.stored[keyFor(1, day1)] = struct{}{}

	// Already stored: same name, existing solve, no solve. Nothing to write.
	b.Observe(day1, []Result{{NYTUserID: 1, Name: "Alice", SolveTimeSeconds: new(int32(30))}})
	b.Observe(day2, []Result{{NYTUserID: 1, Name: "Alice"}})
	// An empty name from the API is not a rename.
	b.Observe(day2, []Result{{NYTUserID: 1, Name: ""}})

	if u, s := b.Pending(); u != 0 || s != 0 {
		t.Fatalf("Pending = %d users, %d solves; want none", u, s)
	}
	if err := b.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 0 {
		t.Fatalf("Flush with nothing pending touched the database: %+v", *calls)
	}
}

func TestBatcherWritesOnlyChanges(t *testing.T) {
	b, calls := recordingBatcher(t, nil)
	b.names[1] = "Alice"
	b.names[2] = "Bob"
	b.stored[keyFor(1, day1)] = struct{}{}

	b.Observe(day1, []Result{
		{NYTUserID: 1, Name: "Alice", SolveTimeSeconds: new(int32(30))}, // stored
		{NYTUserID: 2, Name: "Bob", SolveTimeSeconds: new(int32(50))},   // new solve
		{NYTUserID: 3, Name: "Carol"},                                   // new user
	})
	b.Observe(day2, []Result{
		{NYTUserID: 1, Name: "Alicia", SolveTimeSeconds: new(int32(40))}, // rename + new solve
		{NYTUserID: 2, Name: "Bob", SolveTimeSeconds: new(int32(99))},
	})
	// Seen again before the flush: still one pending solve each, first time kept.
	b.Observe(day1, []Result{{NYTUserID: 2, Name: "Bob", SolveTimeSeconds: new(int32(51))}})

	if err := b.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 1 {
		t.Fatalf("got %d saves, want 1", len(*calls))
	}
	got := (*calls)[0]
	wantUsers := []Result{{NYTUserID: 1, Name: "Alicia"}, {NYTUserID: 2, Name: "Bob"}, {NYTUserID: 3, Name: "Carol"}}
	if !reflect.DeepEqual(got.users, wantUsers) {
		t.Errorf("users = %+v, want %+v", got.users, wantUsers)
	}
	wantSolves := []solve{
		{NYTUserID: 2, Date: day1, Seconds: 50},
		{NYTUserID: 1, Date: day2, Seconds: 40},
		{NYTUserID: 2, Date: day2, Seconds: 99},
	}
	if !reflect.DeepEqual(got.solves, wantSolves) {
		t.Errorf("solves = %+v, want %+v", got.solves, wantSolves)
	}

	// After a successful flush the same observations are no-ops.
	b.Observe(day2, []Result{{NYTUserID: 1, Name: "Alicia", SolveTimeSeconds: new(int32(40))}})
	if u, s := b.Pending(); u != 0 || s != 0 {
		t.Errorf("Pending after flush = %d users, %d solves; want none", u, s)
	}
}

func TestBatcherKeepsChangesWhenFlushFails(t *testing.T) {
	fail := true
	b, calls := recordingBatcher(t, &fail)
	b.Observe(day1, []Result{{NYTUserID: 1, Name: "Alice", SolveTimeSeconds: new(int32(30))}})

	if err := b.Flush(context.Background()); err == nil {
		t.Fatal("Flush succeeded, want error")
	}
	if u, s := b.Pending(); u != 1 || s != 1 {
		t.Fatalf("Pending after failed flush = %d users, %d solves; want 1, 1", u, s)
	}

	fail = false
	if err := b.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 2 || !reflect.DeepEqual((*calls)[0], (*calls)[1]) {
		t.Errorf("retry did not resend the same changes: %+v", *calls)
	}
}

func TestBatcherObserveDuringFlush(t *testing.T) {
	results := []Result{{NYTUserID: 1, Name: "Alice", SolveTimeSeconds: new(int32(30))}}
	var b *Batcher
	saves := 0
	b = newBatcher(func(context.Context, []Result, []solve) error {
		saves++
		// A scrape of the same date lands while the save is in flight.
		b.Observe(day1, results)
		return nil
	})
	b.Observe(day1, results)

	if err := b.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	if u, s := b.Pending(); u != 0 || s != 0 {
		t.Fatalf("Pending = %d users, %d solves; what was just saved was re-queued", u, s)
	}
	if err := b.Flush(context.Background()); err != nil || saves != 1 {
		t.Fatalf("second Flush saved again (saves=%d, err=%v)", saves, err)
	}
}

func TestCookieCache(t *testing.T) {
	loads := 0
	value := "NYT-S=one"
	c := &CookieCache{
		Load: func(context.Context) (string, error) {
			loads++
			return value, nil
		},
		MinReload: time.Hour,
	}
	ctx := context.Background()

	for range 3 {
		if got, err := c.Get(ctx); err != nil || got != "NYT-S=one" {
			t.Fatalf("Get = %q, %v", got, err)
		}
	}
	if loads != 1 {
		t.Fatalf("loaded %d times, want 1", loads)
	}

	// Invalidated, but MinReload has not passed: keep serving the cached value.
	value = "NYT-S=two"
	c.Invalidate()
	if got, _ := c.Get(ctx); got != "NYT-S=one" || loads != 1 {
		t.Fatalf("Get = %q after %d loads; reload should wait for MinReload", got, loads)
	}

	// Once MinReload has passed, the next Get reloads.
	c.attempted = time.Now().Add(-2 * time.Hour)
	if got, _ := c.Get(ctx); got != "NYT-S=two" || loads != 2 {
		t.Fatalf("Get = %q after %d loads; want reload", got, loads)
	}
}

func TestCookieCacheLoadFailure(t *testing.T) {
	c := &CookieCache{
		Load:      func(context.Context) (string, error) { return "", errors.New("db down") },
		MinReload: time.Hour,
	}
	if _, err := c.Get(context.Background()); err == nil {
		t.Fatal("Get succeeded with no cookie")
	}
}
