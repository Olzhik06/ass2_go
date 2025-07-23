package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Server struct {
	mu         sync.Mutex
	data       map[string]string
	requests   int
	shutdownCh chan struct{}
}

func NewServer() *Server {
	return &Server{
		data:       make(map[string]string),
		shutdownCh: make(chan struct{}),
	}
}

func (s *Server) postDataHandler(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	for key, value := range body {
		s.data[key] = value
	}
	s.requests++
	s.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
}
func (s *Server) getDataHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	dataCopy := make(map[string]string)
	for k, v := range s.data {
		dataCopy[k] = v
	}
	s.requests++
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dataCopy)
}

func (s *Server) statsHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	stats := map[string]int{
		"requests":  s.requests,
		"data_size": len(s.data),
	}
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
func (s *Server) deleteDataHandler(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/data/")

	s.mu.Lock()
	if _, exists := s.data[key]; exists {
		delete(s.data, key)
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Key not found", http.StatusNotFound)
	}
	s.requests++
	s.mu.Unlock()
}

func (s *Server) startBackgroundWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		log.Printf("Status: %d requests, %d items in database", s.requests, len(s.data))
		s.mu.Unlock()
	}
}

func (s *Server) shutdown() {
	close(s.shutdownCh)
}
func main() {
	server := NewServer()
	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			server.postDataHandler(w, r)
		} else {
			server.getDataHandler(w, r)
		}
	})

	http.HandleFunc("/stats", server.statsHandler)

	go server.startBackgroundWorker()

	log.Println("Server running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

//type User struct {
//	name string
//	age  int
//}
//
//func home_page(w http.ResponseWriter, r *http.Request) {
//	bob := User{"Bob", 21}
//	fmt.Fprintf(w, "User name and his age: "+bob.name)
//}
//func contacts_page(w http.ResponseWriter, r *http.Request) {}
//
//func HandleRequest() {
//	http.HandleFunc("/", home_page)
//	http.HandleFunc("/contacts/", contacts_page)
//	fmt.Println("Server is active at http://localhost:8080")
//	err := http.ListenAndServe(":8080", nil)
//	if err != nil {
//		fmt.Println("Error has occured", err)
//	}
//
//}
