package ws_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	wsadapter "github.com/blagoySimandov/takgo/internal/adapters/ws"
	"github.com/blagoySimandov/takgo/internal/domain/game"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var testUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func setupPlayerConn(t *testing.T, notifier *wsadapter.WsNotifier, playerID uuid.UUID) *websocket.Conn {
	t.Helper()
	registered := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		notifier.Register(playerID, conn)
		close(registered)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	t.Cleanup(srv.Close)

	client, _, err := websocket.DefaultDialer.Dial("ws"+srv.URL[4:], nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close client: %v", err)
		}
	})

	<-registered
	return client
}

func newFinishedGame(p1, p2 uuid.UUID) *game.Game {
	winner := p1
	return &game.Game{
		ID:      uuid.New(),
		State:   game.Finished,
		Players: [2]game.Player{{ID: p1, Symbol: game.X}, {ID: p2, Symbol: game.O}},
		WinnerID: &winner,
	}
}

func newPlayingGame(p1, p2 uuid.UUID) *game.Game {
	return &game.Game{
		ID:          uuid.New(),
		State:       game.Playing,
		Players:     [2]game.Player{{ID: p1, Symbol: game.X}, {ID: p2, Symbol: game.O}},
		CurrentTurn: p1,
	}
}

func readWithTimeout(t *testing.T, conn *websocket.Conn, timeout time.Duration) ([]byte, error) {
	t.Helper()
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		t.Fatal(err)
	}
	_, msg, err := conn.ReadMessage()
	return msg, err
}

func TestNotifyMoveSendsGameStateToBothPlayers(t *testing.T) {
	notifier := wsadapter.NewWsNotifier()
	p1, p2 := uuid.New(), uuid.New()
	client1 := setupPlayerConn(t, notifier, p1)
	client2 := setupPlayerConn(t, notifier, p2)

	g := newPlayingGame(p1, p2)
	if err := notifier.NotifyMove(context.Background(), g, game.Move{PlayerID: p1, Position: 4}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := readWithTimeout(t, client1, time.Second); err != nil {
		t.Fatalf("p1 expected game state message: %v", err)
	}
	if _, err := readWithTimeout(t, client2, time.Second); err != nil {
		t.Fatalf("p2 expected game state message: %v", err)
	}
}

func TestDeregisteredPlayerDoesNotReceiveNotification(t *testing.T) {
	notifier := wsadapter.NewWsNotifier()
	p1, p2 := uuid.New(), uuid.New()
	client1 := setupPlayerConn(t, notifier, p1)
	client2 := setupPlayerConn(t, notifier, p2)

	notifier.Deregister(p1)

	g := newPlayingGame(p1, p2)
	if err := notifier.NotifyMove(context.Background(), g, game.Move{PlayerID: p2, Position: 0}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := readWithTimeout(t, client1, 100*time.Millisecond); err == nil {
		t.Fatal("expected no message for deregistered player")
	}
	if _, err := readWithTimeout(t, client2, time.Second); err != nil {
		t.Fatalf("p2 expected game state message: %v", err)
	}
}

func TestNotifyMoveOnFinishedGameSendsCloseMessage(t *testing.T) {
	notifier := wsadapter.NewWsNotifier()
	p1, p2 := uuid.New(), uuid.New()
	client := setupPlayerConn(t, notifier, p1)
	setupPlayerConn(t, notifier, p2)

	g := newFinishedGame(p1, p2)
	if err := notifier.NotifyMove(context.Background(), g, game.Move{PlayerID: p1, Position: 2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := readWithTimeout(t, client, time.Second); err != nil {
		t.Fatalf("expected game state before close: %v", err)
	}
	_, err := readWithTimeout(t, client, time.Second)
	if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		t.Fatalf("expected close frame, got: %v", err)
	}
}

func TestNotifyMoveDoesNotSendToUnregisteredPlayer(t *testing.T) {
	notifier := wsadapter.NewWsNotifier()
	p1, p2 := uuid.New(), uuid.New()
	client2 := setupPlayerConn(t, notifier, p2)

	g := newPlayingGame(p1, p2)
	if err := notifier.NotifyMove(context.Background(), g, game.Move{PlayerID: p1, Position: 0}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := readWithTimeout(t, client2, time.Second); err != nil {
		t.Fatalf("p2 expected game state message: %v", err)
	}
}
