package main

import (
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

/*
 * Structs
 */
type Email struct {
	From    string
	To      string
	Subject string
	Body    string
}

/*
 * Sync Pool
 */
var EmailPool = sync.Pool{
	New: func() interface{} {
		return new(Email)
	},
}
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 64*1024)
	},
}

var filePool = sync.Pool{
	New: func() interface{} {
		return new(os.File)
	},
}

/*
 * Variables
 */
var data []map[string]interface{}
var bufferSize = 64 * 1024

const (
	url      = "http://localhost:4080/api/_bulk"
	username = "admin"
	password = "Complexpass#123"
)

func main() {
	f, err := os.Create("log/cpu.prof")
	if err != nil {
		fmt.Printf("No se pudo crear el archivo err: %v", err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	root := os.Args[1] //path
	var wg sync.WaitGroup

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			wg.Add(1)
			go readByByte(path, &wg)
			defer wg.Done()

		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error al recorrer la ruta%q: %v\n", root, err)
	}

	wg.Wait()
	defer convertAndSend()
}

func readByByte(path string, wg *sync.WaitGroup) error {
	line := []byte{}
	var body []string

	defer wg.Done()
	file := filePool.Get().(*os.File)
	defer filePool.Put(file)

	var err error
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	buffer := bufferPool.Get().([]byte)
	defer bufferPool.Put(buffer)

	for {
		n, err := f.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)

			}
			if n == 0 {
				break
			}
		}

		//Se crean las lineas hasta el salto de linea
		for _, b := range buffer[:n] {
			line = append(line, b)
			if b == '\n' {

				body = append(body, string(line))
				line = []byte{}
			}
		}
	}
	if len(data) > 15 {

		err = emailFormat(body)
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func emailFormat(lines []string) error {
	newEmail := EmailPool.Get().(*Email)
	defer EmailPool.Put(&newEmail)

	newEmail.From = strings.Split(lines[2], "From:")[1]
	newEmail.To = strings.Join(strings.Split(lines[3], "To:")[1:], "")
	newEmail.Subject = strings.Join(strings.Split(lines[4], "Subject:")[1:], "")
	newEmail.Body = strings.Join(lines[15:], "")

	data = append(data, map[string]interface{}{
		"index": map[string]interface{}{"_index": "enron_test"},
	}, map[string]interface{}{
		"from": newEmail.From, "to": newEmail.To, "subject": newEmail.Subject, "body": newEmail.Body,
	})
	return nil
}

func convertAndSend() {
	fmt.Println("Convert...")
	file, err := os.Create("data/data.ndjson")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	for _, d := range data {
		enc.Encode(d)

		if err != nil {
			fmt.Println(err)
			continue
		}

	}
	defer sendFileChunk()
}

func sendFileChunk() {
	fmt.Println("Sending...")

	f := filePool.Get().(*os.File)
	defer filePool.Put(f)

	file, err := os.Open("data/data.ndjson")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Obtener el tamaño del archivo
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	fileSize := fileInfo.Size()

	// Crear un buffer para leer el archivo
	buffer := make([]byte, 1024*1024)

	// Crear una nueva solicitud HTTP POST
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(username, password)
	var start int64
	for start < fileSize {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			panic(err)
		}
		end := start + int64(n)

		req.Body = ioutil.NopCloser(io.NewSectionReader(file, start, end-start))
		err = req.Body.Close()
		if err != nil {
			return
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		start = end
		defer resp.Body.Close()
	}
}
