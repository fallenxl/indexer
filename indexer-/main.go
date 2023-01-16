package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"sync"
)

type Email struct {
	From    string
	To      string
	Subject string
	Body    string
}

var data []map[string]interface{}

const url = "http://localhost:4080/api/_bulk"
const username = "admin"
const password = "Complexpass#123"

func main() {
	f, e := os.Create("log/cpu.prof")
	if e != nil {
		fmt.Printf("Cannot create the file err: %v", e)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	root := "enron_mail/maildir/allen-p/_sent_mail/9" //path
	var wg sync.WaitGroup
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// email := readByByte(path, 1024*100)
				email := readByLine(path)
				data = append(data, map[string]interface{}{
					"index": map[string]interface{}{"_index": "enron_mail"},
				}, map[string]interface{}{
					"from": email.From, "to": email.To, "subject": email.Subject, "body": email.Body,
				})

			}()
		}
		return nil
	})
	wg.Wait()
	defer convertAndSend()
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", root, err)
	}
}

func readByByte(path string, chunkSize uint64) Email {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)

	}
	defer file.Close()

	buf := make([]byte, chunkSize)
	line := []byte{}
	var data []string
	var email Email
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)

			}
			if n == 0 {
				break
			}
		}

		for _, b := range buf[:n] {
			line = append(line, b)
			if b == '\n' {
				// do something with the line
				data = append(data, string(line))
				line = []byte{}
			}
		}
	}

	email = Email{
		From:    strings.Split(data[2], "From:")[1],
		To:      strings.Join(strings.Split(data[3], "To:")[1:], ""),
		Subject: strings.Join(strings.Split(data[4], "Subject:")[1:], ""),
		Body:    strings.Join(data[15:], ""),
	}

	return email

}

func readByLine(path string) Email {
	var data []string
	var email Email
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("error reading the file %q: %v\n", path, err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		text := strings.Split(string(line), ":")
		data = append(data, text...)
	}
	if len(data) > 32 {
		body := strings.Join(data[33:], "")
		email = Email{
			From:    data[7],
			To:      data[9],
			Subject: data[11],
			Body:    body,
		}
	}
	return email
}

func convertAndSend() {
	file, err := os.Create("data.ndjson")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for _, d := range data {
		jsonData, err := json.Marshal(d)
		if err != nil {
			fmt.Println(err)
			return
		}

		_, err = file.Write(jsonData)
		if err != nil {
			fmt.Println(err)
			return
		}
		file.Write([]byte{'\n'})

	}
	defer indexer()
}

func indexer() {
	file, err := os.Open("data.ndjson")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	jsonData, err := ioutil.ReadAll(file)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(username, password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
