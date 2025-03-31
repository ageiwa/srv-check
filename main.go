package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Server struct {
	Uri string `json:"uri"`
	StatusCode int `json:"statusCode"`
	ResponseTime int `json:"responseTime"`
}

const SERVER_FILE = "server.json"
const PERM_CODE = 0600

const QUIT = "q"
const ADD = "add"
const LIST = "ls"
const DELETE = "del"

var gFile *os.File

func madeRequest() {
	// startTime := time.Now()
	resp, err := http.Get("https://aiostudy.com")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer resp.Body.Close()

	// resp.StatusCode
	// time.Since(startTime).Milliseconds()
}

func checkout() {
	data, err := os.ReadFile(SERVER_FILE)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("No servers for checking...")
		} else {
			fmt.Println("Unknown file reading error...")
		}

		return
	}

	fmt.Println(data)
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
	servers := []Server{}

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

	fmt.Println(servers)

	newServer := Server{
		Uri: serverName,
	}

	servers = append(servers, newServer)

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
		fmt.Println("Ошибка записи в файл")
		return
	}

	fmt.Println("Server was added...")
}

func removeServer(serverName string) {
	servers := []Server{}

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
		if server.Uri == serverName {
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
		fmt.Println("Не удалось записать файл")
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

		// for {
		// 	fmt.Println("Checking...")

		// 	checkout()

		// 	timer := time.NewTimer(time.Duration(5 * time.Second))
		// 	<-timer.C
		// }
	}()

	wg.Wait()
}