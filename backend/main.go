package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Container struct {
	Name string `json:"name"`
	CPU  string `json:"cpu"`
	RAM  string `json:"ram"`
	Logs string `json:"logs"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan []Container)
	mu        sync.Mutex
)

func getContainers() []Container {
	cmd := exec.Command("docker", "stats", "--no-stream", "--format", "{{.Name}};{{.CPUPerc}};{{.MemUsage}}")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Println("Erreur récupération des stats:", err)
		return nil
	}
	var containers []Container
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ";")
		containers = append(containers, Container{Name: parts[0], CPU: parts[1], RAM: parts[2]})
	}
	return containers
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Erreur de connexion WebSocket :", err)
		return
	}
	defer conn.Close()
	mu.Lock()
	clients[conn] = true
	mu.Unlock()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			break
		}
	}
}

func getContainerLogs(containerName string) string {
	cmd := exec.Command("docker", "logs", "--tail", "10", containerName)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Println("Erreur récupération logs:", err)
		return "Erreur récupération logs"
	}
	return out.String()
}

func broadcastContainers() {
	for {
		containers := getContainers()
		if len(containers) == 0 {
			continue
		}
		for i := range containers {
			containers[i].Logs = getContainerLogs(containers[i].Name)
		}
		data, err := json.Marshal(containers)
		if err != nil {
			log.Println("Erreur encodage JSON:", err)
			continue
		}
		mu.Lock()
		for conn := range clients {
			err := conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println("Erreur envoi message :", err)
				conn.Close()
				delete(clients, conn)
			}
		}
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	go broadcastContainers()
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Erreur serveur :", err)
	}
}