package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"sync"

	"github.com/edsrzf/mmap-go"
)

type Email struct {
	From    string "json:from"
	To      string "json:to"
	Subject string "json:subject"
	Body    string "json:body"
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

	root := os.Args[1] //path
	var wg sync.WaitGroup

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			wg.Add(1)

			go func() {
				defer wg.Done()
				email := readByLine(path)
				data = append(data, map[string]interface{}{
					"index": map[string]interface{}{"_index": "test"},
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

func test(path string, wg *sync.WaitGroup) {
	defer wg.Done()
	line := []byte{}
	var data []string
	buffer := make([]byte, 32*1024*1024)
	var filePool = sync.Pool{
		New: func() interface{} {
			f, _ := os.Open(path)
			return f
		},
	}

	f := filePool.Get().(*os.File)
	defer filePool.Put(f)
	for {
		n, err := f.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}
		if n == 0 {
			break
		}
		// Process buffer here
		for _, b := range buffer[:n] {
			line = append(line, b)
			if b == '\n' {
				// do something with the line
				data = append(data, string(line))
				line = []byte{}
			}
		}
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

func read(path string) Email {
	var email Email
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	b, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		fmt.Println(err)
	}
	defer b.Unmap()

	// Do something with the content of the file
	data := strings.Split(string(b), ":")
	if len(data) > 17 {
		email = Email{
			From:    strings.Split(data[5], "\n")[0],
			To:      strings.Split(data[6], "\n")[0],
			Subject: strings.Split(data[7], "\n")[0],
			Body:    strings.Join(strings.Split(data[17], "\n")[1:], ""),
		}

	}

	return email
}

func readByLine(path string) Email {
	var data []string
	var email Email
	var filePool = sync.Pool{
		New: func() interface{} {
			f, _ := os.Open(path)
			return f
		},
	}

	f := filePool.Get().(*os.File)
	defer filePool.Put(f)

	reader := bufio.NewReader(f)
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

	enc := json.NewEncoder(file)
	for _, d := range data {
		enc.Encode(d)
		if err != nil {
			fmt.Println(err, "fallo en el encode")
			continue
		}

	}
	defer indexer("data.ndjson")
}

func indexer(data string) {
	file, err := os.Open(data)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	f, err := bufio.NewReader(file).ReadString('\n')
	if err != nil {
		fmt.Println("Fallo en el bufio.newreader")
	}
	req, err := http.NewRequest("POST", url, strings.NewReader(string(f)))
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

func indexOneByOne(data Email) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}
	req, err := http.NewRequest("POST", "http://localhost:4080/api/test/_doc", strings.NewReader(string(jsonData)))
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth("admin", "Complexpass#123")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	// log.Println(resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(body)
}
