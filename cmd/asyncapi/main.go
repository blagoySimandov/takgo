package main

import (
	"log"
	"os"

	wsadapter "github.com/blagoySimandov/takgo/internal/adapters/ws"
	asyncapi "github.com/swaggest/go-asyncapi/reflector/asyncapi-2.4.0"
	spec "github.com/swaggest/go-asyncapi/spec-2.4.0"
)

func main() {
	schema := &spec.AsyncAPI{}
	schema.Info.Title = "TakGo WebSocket API"
	schema.Info.Version = "1.0.0"
	schema.Info.Description = "Real-time tic-tac-toe over WebSocket. Connect to be matched, then exchange moves."
	schema.AddServer("production", spec.Server{
		URL:      "localhost:8080",
		Protocol: "ws",
	})

	r := asyncapi.Reflector{Schema: schema}
	mustAdd(r.AddChannel(asyncapi.ChannelInfo{
		Name: "/api/v1/game/connect",
		Publish: &asyncapi.MessageSample{
			MessageEntity: spec.MessageEntity{
				Summary:     "Make a move",
				Description: "Client sends a move. Rejected if out of turn, cell occupied, or game finished.",
			},
			MessageSample: new(wsadapter.MoveMsg),
		},
		Subscribe: &asyncapi.MessageSample{
			MessageEntity: spec.MessageEntity{
				Summary:     "Game state update",
				Description: "Server pushes updated board state after every move. On game over a WS close frame follows.",
			},
			MessageSample: new(wsadapter.GameStateMsg),
		},
	}))

	yaml, err := r.Schema.MarshalYAML()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("asyncapi.yaml", yaml, 0644); err != nil {
		log.Fatal(err)
	}
	log.Println("written asyncapi.yaml")
}

func mustAdd(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
