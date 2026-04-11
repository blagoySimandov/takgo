package matchmaking

import (
	"time"

	"github.com/blagoySimandov/takgo/internal/domain/game"
	"github.com/google/uuid"
)

type Queue struct {
	waiting chan uuid.UUID
}

func NewQueue() *Queue {
	// we could make this buffered and support multiple players....
	// we would also need to "spawn" a q goroutine to handle the queue and try to pair people continuously
	// and Join will no longer return a game....
	// you would need to listen to a channel of pairs and create a game conteniously..
	// yeah fuck that.
	return &Queue{waiting: make(chan uuid.UUID, 1)}
}

// Join adds a player to the queue. Returns a new game when two players are
// matched, or nil if the player is now waiting.
func (q *Queue) Join(playerID uuid.UUID) *game.Game {
	select {
	case opponent := <-q.waiting:
		return newGame(opponent, playerID)
	case q.waiting <- playerID:
		return nil
	}
}

// Leave removes a waiting player from the queue. Returns false if the player
// was not in the queue.
func (q *Queue) Leave(playerID uuid.UUID) bool {
	select {
	case id := <-q.waiting:
		if id == playerID {
			return true
		}
		q.waiting <- id // not our player, put them back
		// ahh... this could be problematic if we have more than 2 players..
		// a mutex and a slice would be better so the order is preserved....
		// Anyway just pointing things out
		return false
	default:
		return false
	}
}

func newGame(first, second uuid.UUID) *game.Game {
	return &game.Game{
		ID:    uuid.New(),
		State: game.Playing,
		Players: [2]game.Player{
			{ID: first, Symbol: game.X},
			{ID: second, Symbol: game.O},
		},
		CurrentTurn: first,
		CreatedAt:   time.Now(),
	}
}
