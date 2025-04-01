package main

import (
	"bufio"
	"encoding/json"
	"errors"

	// "errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Uri          string `json:"uri"`
	StatusCode   int    `json:"statusCode"`
	ResponseTime int    `json:"responseTime"`
}

type RequestInfo struct {
	StatusCode   int `json:"statusCode"`
	ResponseTime int64 `json:"responseTime"`
}

const SERVER_FILE = "servers.json"
const PERM_CODE = 0600

const QUIT = "q"
const ADD = "add"
const LIST = "ls"
const DELETE = "del"

var gFile *os.File

func madeRequest(uri string) (RequestInfo, error) {
	startTime := time.Now()
	resp, err := http.Get(uri)

	if err != nil {
		if errors.Is(err, http.ErrNotSupported) {
			fmt.Println("Ошибка. это не поддерживается")
		}

		return RequestInfo{}, err
	}

	defer resp.Body.Close()

	return RequestInfo{
		StatusCode: resp.StatusCode,
		ResponseTime: time.Since(startTime).Milliseconds(),
	}, nil
}

func checkout() {
	servers := []string{}

	if err := json.NewDecoder(gFile).Decode(&servers); err != nil {
		if err != io.EOF {
			fmt.Println(err.Error())
			return
		}
	}

	if len(servers) == 0 {
		fmt.Println("No servers for checking...")
		return
	}

	for _, server := range servers {
		info, err := madeRequest(server)

		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		fmt.Println("status code:", info.StatusCode)
		fmt.Println("response time:", info.ResponseTime)
	}

	if _, err := gFile.Seek(0, 0); err != nil {
		fmt.Println("Can't move pointer")
	}
}

func clearFile() error {
	if err := gFile.Truncate(0); err != nil {
		return err
	}

	if _, err := gFile.Seek(0, 0); err != nil {
		return err
	}

	return nil
}

func addServerForChecking(serverName string) {
	servers := []string{}

	if _, err := gFile.Seek(0, 0); err != nil {
		fmt.Println("Can't move pointer")
		return
	}

	if err := json.NewDecoder(gFile).Decode(&servers); err != nil {
		if err != io.EOF {
			fmt.Println(err.Error())
			return
		}
	}

	for _, server := range servers {
		if server == serverName {
			fmt.Println("Server already exists...")
			return
		}
	}

	servers = append(servers, serverName)
	data, err := json.Marshal(servers)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if err := clearFile(); err != nil {
		fmt.Println("Can't clear file")
		return
	}

	if _, err := gFile.Write(data); err != nil {
		fmt.Println("Can't write to file")
		return
	}

	fmt.Println("Server was added...")
}

func removeServer(serverName string) {
	servers := []string{}

	if _, err := gFile.Seek(0, 0); err != nil {
		fmt.Println("Can't move pointer")
		return
	}

	if err := json.NewDecoder(gFile).Decode(&servers); err != nil {
		if err != io.EOF {
			fmt.Println(err.Error())
			return
		}
	}

	for i, server := range servers {
		if server == serverName {
			servers = append(servers[:i], servers[i+1:]...)
			break
		}
	}

	data, err := json.Marshal(servers)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if err := clearFile(); err != nil {
		fmt.Println("Can't clear file")
		return
	}

	if _, err := gFile.Write(data); err != nil {
		fmt.Println("Can't write to file")
		return
	}

	fmt.Println(serverName, "was removed")
}

func executeCmd(key string, args []string) {
	switch key {
	case QUIT:
		os.Exit(0)
		fmt.Println("Stopping")
	case ADD:
		addServerForChecking(args[0])
	case DELETE:
		removeServer(args[0])
	default:
		fmt.Println("Unknown command")
	}
}

func main() {
	file, err := os.OpenFile(SERVER_FILE, os.O_RDWR|os.O_CREATE, PERM_CODE)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	gFile = file

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

		for {
			reader := bufio.NewReader(os.Stdin)
			cmd, err := reader.ReadString('\n')

			if err != nil {
				err := fmt.Errorf("Error reading command...")
				fmt.Println(err)
				continue
			}

			cmd = strings.TrimSpace(cmd)
			cmds := strings.Split(cmd, " ")

			executeCmd(cmds[0], cmds[1:])
		}
	}()

	go func() {
		defer wg.Done()

		for {
			fmt.Println("Checking...")

			checkout()

			timer := time.NewTimer(time.Duration(5 * time.Second))
			<-timer.C
		}
	}()

	wg.Wait()
}
