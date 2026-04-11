package game_test

import (
	"context"
	"testing"

	"github.com/blagoySimandov/takgo/internal/domain/game"
	"github.com/google/uuid"
)

// fakeQueue implements game.IQueue for hub tests.
type fakeQueue struct {
	joined  []uuid.UUID
	left    []uuid.UUID
	matched *game.Game
}

func (q *fakeQueue) Join(playerID uuid.UUID) *game.Game {
	q.joined = append(q.joined, playerID)
	return q.matched
}

func (q *fakeQueue) Leave(playerID uuid.UUID) bool {
	q.left = append(q.left, playerID)
	return true
}

// fakeNotifier records calls to NotifyMove.
type fakeNotifier struct {
	calls int
}

func (n *fakeNotifier) NotifyMove(_ context.Context, _ *game.Game, _ game.Move) error {
	n.calls++
	return nil
}

func newTestHub(q game.IQueue, notifier game.Notifier) *game.Hub {
	return game.NewHub(q, game.NewInMemoryGameRepo(), notifier)
}

func TestConnectingToEmptyQueueReturnsPendingChannel(t *testing.T) {
	q := &fakeQueue{}
	hub := newTestHub(q, &fakeNotifier{})

	ch := hub.Connect(uuid.New())

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected empty channel, got value")
		}
	default:
		// channel open and empty — correct
	}
}

func TestConnectWhenMatchedImmediatelyDeliversGame(t *testing.T) {
	p1, p2 := uuid.New(), uuid.New()
	g := &game.Game{
		ID:      uuid.New(),
		State:   game.Playing,
		Players: [2]game.Player{{ID: p1, Symbol: game.X}, {ID: p2, Symbol: game.O}},
		CurrentTurn: p1,
	}
	q := &fakeQueue{matched: g}
	notifier := &fakeNotifier{}
	hub := newTestHub(q, notifier)

	hub.Connect(p2) // pre-register p2 so notifyOpponent finds them
	ch := hub.Connect(p1)

	select {
	case got, ok := <-ch:
		if !ok {
			t.Fatal("channel closed unexpectedly")
		}
		if got.ID != g.ID {
			t.Fatalf("expected game %v, got %v", g.ID, got.ID)
		}
	default:
		t.Fatal("expected game on channel, got nothing")
	}
}

func TestDisconnectClosesPendingChannel(t *testing.T) {
	q := &fakeQueue{}
	hub := newTestHub(q, &fakeNotifier{})

	playerID := uuid.New()
	ch := hub.Connect(playerID)
	hub.Disconnect(playerID)

	_, ok := <-ch
	if ok {
		t.Fatal("expected channel to be closed after Disconnect")
	}
}

func TestDisconnectCallsQueueLeave(t *testing.T) {
	q := &fakeQueue{}
	hub := newTestHub(q, &fakeNotifier{})

	playerID := uuid.New()
	hub.Connect(playerID)
	hub.Disconnect(playerID)

	if len(q.left) != 1 || q.left[0] != playerID {
		t.Fatalf("expected Leave(%v), got %v", playerID, q.left)
	}
}

func TestMakeMoveReturnsErrorForUnknownGame(t *testing.T) {
	hub := newTestHub(&fakeQueue{}, &fakeNotifier{})

	err := hub.MakeMove(context.Background(), uuid.New(), uuid.New(), 0)

	if err == nil {
		t.Fatal("expected error for unknown game, got nil")
	}
}

func TestMakeMoveNotifiesOnSuccess(t *testing.T) {
	p1, p2 := uuid.New(), uuid.New()
	g := &game.Game{
		ID:    uuid.New(),
		State: game.Playing,
		Players: [2]game.Player{
			{ID: p1, Symbol: game.X},
			{ID: p2, Symbol: game.O},
		},
		CurrentTurn: p1,
	}
	repo := game.NewInMemoryGameRepo()
	_ = repo.Create(context.Background(), g)
	notifier := &fakeNotifier{}
	hub := game.NewHub(&fakeQueue{}, repo, notifier)

	if err := hub.MakeMove(context.Background(), g.ID, p1, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notifier.calls != 1 {
		t.Fatalf("expected 1 notifier call, got %d", notifier.calls)
	}
}

func TestMakeMoveReturnsErrNotYourTurn(t *testing.T) {
	p1, p2 := uuid.New(), uuid.New()
	g := &game.Game{
		ID:    uuid.New(),
		State: game.Playing,
		Players: [2]game.Player{
			{ID: p1, Symbol: game.X},
			{ID: p2, Symbol: game.O},
		},
		CurrentTurn: p1,
	}
	repo := game.NewInMemoryGameRepo()
	_ = repo.Create(context.Background(), g)
	hub := game.NewHub(&fakeQueue{}, repo, &fakeNotifier{})

	err := hub.MakeMove(context.Background(), g.ID, p2, 0)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
