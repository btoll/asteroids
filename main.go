package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func handleConnection(ctx context.Context, gameServer *GameServer, conn net.Conn) {
	select {
	case <-ctx.Done():
		conn.Close()
		return
	default:
	}

	conn.Write([]byte("What is your name?: "))
	scanner := bufio.NewScanner(conn)
	var client *Client
	for scanner.Scan() {
		name := scanner.Text()
		gameServer.mu.Lock()
		found := gameServer.IsRegistered(name)
		gameServer.mu.Unlock()
		if found {
			fmt.Fprintf(conn, "The username %s is already taken, choose another: ", name)
		} else {
			fmt.Fprintf(conn, "Hi %s, welcome to the game!\n", name)
			gameServer.mu.Lock()
			client = NewClient(Username(name), conn)
			gameServer.Register(client)
			gameServer.mu.Unlock()
			break
		}
	}
	err := scanner.Err()
	if !errors.Is(err, io.EOF) && err != nil {
		log.Printf("[ERROR] %s\n", err.Error())
	}
	client.Start(ctx)
}

func shutdown(gameServer *GameServer, cancel context.CancelFunc) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	gameServer.mu.Lock()
	cancel()
	err := gameServer.Shutdown()
	if err != nil {
		log.Printf("[ERROR] %s\n", err.Error())
	}
	gameServer.mu.Unlock()
}

func main() {
	listener, err := net.Listen("tcp", ":8111")
	if err != nil {
		log.Printf("[ERROR] %s\n", err.Error())
		return
	}
	log.Printf("[INFO] %s server listening on %s...\n", listener.Addr().Network(), listener.Addr())

	gameServer := NewGameServer(listener)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("you're done, fella!")
				return
			}
		}
	}()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Printf("net.ErrClosed=%+v\n", net.ErrClosed)
				if errors.Is(err, net.ErrClosed) {
					log.Println("[INFO] The accept loop is shutting down.")
					return
				}
				log.Printf("[ERROR] %s\n", err.Error())
				continue
			}
			go handleConnection(ctx, gameServer, conn)
		}
	}()

	shutdown(gameServer, cancel)
}
