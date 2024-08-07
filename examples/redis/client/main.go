package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type Response struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Uptime string `json:"uptime"`
	ConnectedClients string `json:"connected_clients"`
}

func clearConsole() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func fetchData(url string) (*Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned error: %s", string(body))
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}

func main() {
	serverURL := "http://localhost:8080/get?key=counter"

	for {
		response, err := fetchData(serverURL)
		clearConsole()
		if err != nil {
			log.Printf("Error fetching data: %v", err)
		} else {
			fmt.Printf("CLIENT\n-----------\nKey: %s\nValue: %s\nUptime: %s seconds\nConnected clients: %s\n", response.Key, response.Value, response.Uptime, response.ConnectedClients)
		}

		time.Sleep(1 * time.Second)
	}
}
