package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var port int
	var host string

	flag.IntVar(&port, "port", 8080, "Port to listen on")
	flag.StringVar(&host, "host", "localhost", "Host to bind to")
	flag.Parse()

	server, err := NewMockServer(host, port)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.HandleWebSocket)
	setupAPI(mux, server)

	server.SetHandlerFunc(mux)

	fmt.Printf("Mock server starting on %s:%d\n", host, port)
	fmt.Printf("WebSocket endpoint: ws://%s:%d/\n", host, port)
	fmt.Printf("REST API endpoint: http://%s:%d/api/\n", host, port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-sigChan
	fmt.Println("\nShutting down server...")
	if err := server.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}
	fmt.Println("Server stopped")
}
