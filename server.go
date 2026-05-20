package main

import (
	"log"
	"net"
	"sync"
)

type GameServer struct {
	connections map[Username]*Client
	broadcast   chan struct{}
	listener    net.Listener
	mu          sync.Mutex
}

func NewGameServer(listener net.Listener) *GameServer {
	return &GameServer{
		connections: make(map[Username]*Client),
		listener:    listener,
		mu:          sync.Mutex{},
	}
}

func (g *GameServer) IsRegistered(name string) bool {
	_, found := g.connections[Username(name)]
	return found
}

func (g *GameServer) Register(client *Client) {
	g.connections[client.name] = client
}

func (g *GameServer) Unregister(client *Client) {
	delete(g.connections, client.name)
}

func (g *GameServer) Shutdown() error {
	// Build a snapshot first so we're not ranging over the connections
	// and deleting then at the same time (undefined behavior).
	clients := make([]*Client, 0, len(g.connections))
	for _, client := range g.connections {
		clients = append(clients, client)
	}
	for _, client := range clients {
		err := client.Close()
		if err != nil {
			log.Printf("[ERROR] %s\n", err.Error())
		}
		g.Unregister(client)
	}
	return g.listener.Close()
}
