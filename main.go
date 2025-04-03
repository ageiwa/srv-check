package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	Uri          string `json:"uri"`
	StatusCode   int    `json:"statusCode"`
	ResponseTime int64  `json:"responseTime"`
	Message      string `json:"message"`
}

const SERVER_FILE = "servers.json"
const PERM_CODE = 0600

const QUIT = "q"
const ADD = "add"
const LIST = "ls"
const DELETE = "del"

var gFile *os.File

func madeRequest(uri string) RequestInfo {
	startTime := time.Now()
	resp, err := http.Get(uri)

	if err != nil {
		info := RequestInfo{}
		info.Uri = uri

		if _, ok := err.(*url.Error); ok {
			info.Message = "Invalid url"
		} else {
			info.Message = "Unknow error"
		}

		return info
	}

	defer resp.Body.Close()

	return RequestInfo{
		Uri:          uri,
		StatusCode:   resp.StatusCode,
		ResponseTime: time.Since(startTime).Milliseconds(),
	}
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

	servsersInfo := []RequestInfo{}

	for _, server := range servers {
		info := madeRequest(server)
		servsersInfo = append(servsersInfo, info)
	}

	dateNow := time.Now()
	newPath := filepath.Join(".", "logs", dateNow.Format(time.DateOnly))

	if _, err := os.Stat(newPath); err != nil {
		if err := os.MkdirAll(newPath, PERM_CODE); err != nil {
			fmt.Println(err.Error(), "Can't create folder")
			return
		}
	}

	fileName := fmt.Sprintf("%s.json", dateNow.Format("15-04-05"))
	file, err := os.OpenFile(filepath.Join(newPath, fileName), os.O_CREATE|os.O_WRONLY, PERM_CODE)

	if err != nil {
		fmt.Println("Can't create file...")
		return
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")

	if err := encoder.Encode(servsersInfo); err != nil {
		fmt.Println("Can't write to file...")
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
