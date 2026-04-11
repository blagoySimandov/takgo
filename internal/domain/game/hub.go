package game

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
)

// IQueue abstracts matchmaking to avoid a circular import (matchmaking imports game).
type IQueue interface {
	Join(playerID uuid.UUID) *Game
	Leave(playerID uuid.UUID) bool
}

type Hub struct {
	queue   IQueue
	gameSvc *GameService
	repo    GameRepository
	pending sync.Map // uuid.UUID -> chan *Game
}

func NewHub(queue IQueue, repo GameRepository, notifier Notifier) *Hub {
	return &Hub{
		queue:   queue,
		repo:    repo,
		gameSvc: NewGameService(repo, notifier),
	}
}

// Connect registers the player in the queue and returns a channel that
// receives the matched game. The pending entry is stored before joining
// the queue so the opponent can always find it.
func (h *Hub) Connect(playerID uuid.UUID) <-chan *Game {
	ch := make(chan *Game, 1)
	h.pending.Store(playerID, ch)

	g := h.queue.Join(playerID)
	if g == nil {
		return ch
	}
	h.pending.Delete(playerID)
	h.repo.Create(context.Background(), g)
	h.notifyOpponent(g, playerID)
	ch <- g
	return ch
}

// Disconnect removes the player from the queue and closes their pending channel.
func (h *Hub) Disconnect(playerID uuid.UUID) {
	h.queue.Leave(playerID)
	if v, ok := h.pending.LoadAndDelete(playerID); ok {
		close(v.(chan *Game))
	}
}

func (h *Hub) MakeMove(ctx context.Context, gameID, playerID uuid.UUID, position int) error {
	return h.gameSvc.MakeMove(ctx, gameID, playerID, position)
}

func (h *Hub) notifyOpponent(g *Game, self uuid.UUID) {
	for _, p := range g.Players {
		if p.ID != self {
			if v, ok := h.pending.Load(p.ID); ok {
				v.(chan *Game) <- g
			}
			return
		}
	}
}

// NewInMemoryGameRepo returns an in-memory GameRepository for active games.
func NewInMemoryGameRepo() GameRepository {
	return &inMemoryGameRepo{games: make(map[uuid.UUID]*Game)}
}

type inMemoryGameRepo struct {
	mu    sync.RWMutex
	games map[uuid.UUID]*Game
}

func (r *inMemoryGameRepo) Create(_ context.Context, g *Game) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.games[g.ID] = g
	return nil
}

func (r *inMemoryGameRepo) FindByID(_ context.Context, id uuid.UUID) (*Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	g, ok := r.games[id]
	if !ok {
		return nil, errors.New("game not found")
	}
	return g, nil
}

func (r *inMemoryGameRepo) Save(_ context.Context, g *Game) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.games[g.ID] = g
	return nil
}
