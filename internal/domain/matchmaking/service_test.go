package matchmaking_test

import (
	"sync"
	"testing"

	"github.com/blagoySimandov/takgo/internal/domain/game"
	"github.com/blagoySimandov/takgo/internal/domain/matchmaking"
	"github.com/google/uuid"
)

func TestPlayerJoiningEmptyQueueWaits(t *testing.T) {
	q := matchmaking.NewQueue()

	result := q.Join(uuid.New())

	if result != nil {
		t.Fatal("expected nil game while waiting")
	}
}

func TestTwoPlayersGetMatched(t *testing.T) {
	q := matchmaking.NewQueue()
	p1, p2 := uuid.New(), uuid.New()

	q.Join(p1)
	g := q.Join(p2)

	if g == nil {
		t.Fatal("expected game to be created")
	}
}

func TestMatchedGameIsInPlayingState(t *testing.T) {
	q := matchmaking.NewQueue()
	p1, p2 := uuid.New(), uuid.New()

	q.Join(p1)
	g := q.Join(p2)

	if g.State != game.Playing {
		t.Fatalf("expected Playing state, got %v", g.State)
	}
}

func TestMatchedPlayersHaveDifferentSymbols(t *testing.T) {
	q := matchmaking.NewQueue()
	p1, p2 := uuid.New(), uuid.New()

	q.Join(p1)
	g := q.Join(p2)

	if g.Players[0].Symbol == g.Players[1].Symbol {
		t.Fatal("expected players to have different symbols")
	}
}

func TestFirstPlayerGetsFirstTurn(t *testing.T) {
	q := matchmaking.NewQueue()
	p1, p2 := uuid.New(), uuid.New()

	q.Join(p1)
	g := q.Join(p2)

	if g.CurrentTurn != p1 {
		t.Fatal("expected first player to have first turn")
	}
}

func TestQueueAcceptsNewPlayerAfterMatch(t *testing.T) {
	q := matchmaking.NewQueue()
	p1, p2, p3 := uuid.New(), uuid.New(), uuid.New()

	q.Join(p1)
	q.Join(p2)
	result := q.Join(p3)

	if result != nil {
		t.Fatal("expected p3 to wait, not match immediately")
	}
}

func TestPlayerCanLeaveQueue(t *testing.T) {
	q := matchmaking.NewQueue()
	p1 := uuid.New()

	q.Join(p1)
	removed := q.Leave(p1)

	if !removed {
		t.Fatal("expected player to be removed")
	}
}

func TestLeaveReturnsFalseIfPlayerNotInQueue(t *testing.T) {
	q := matchmaking.NewQueue()

	removed := q.Leave(uuid.New())

	if removed {
		t.Fatal("expected false for player not in queue")
	}
}

func TestConcurrentJoinsProduceSingleGame(t *testing.T) {
	q := matchmaking.NewQueue()
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		games []*game.Game
	)

	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if g := q.Join(uuid.New()); g != nil {
				mu.Lock()
				games = append(games, g)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if len(games) != 5 {
		t.Fatalf("expected 5 games from 10 players, got %d", len(games))
	}
}
