package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
)

const PORT = ":4000"

func main() {
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r := chi.NewRouter()
	r.Use(cors.Handler)
	r.Use(middleware.Logger)

	r.Get("/{term}", getDocument)
	r.Get("/", getAllDocuments)

	fmt.Printf("Server is running on port %v", PORT)
	err := http.ListenAndServe(PORT, r)
	if err != nil {
		log.Fatal(err)
	}

}

func getAllDocuments(w http.ResponseWriter, r *http.Request) {
	body, err := fetchApi("", "alldocuments")
	if err != nil {
		log.Fatal(err)
	}
	render.JSON(w, r, body)
}

func getDocument(w http.ResponseWriter, r *http.Request) {
	term := chi.URLParam(r, "term")
	body, err := fetchApi(term, "term")
	if err != nil {
		log.Fatal(err)
	}

	render.JSON(w, r, body)
}

func query(str string, t string) (string, error) {
	body := fmt.Sprintf(`{"search_type": "%s",
        "query":
        {
            "term": "%s"
        },
        "from": 0,
        "max_results": 30,
        "_source": []
    }`, t, str)
	fmt.Println(body)
	return body, nil
}

func fetchApi(str string, t string) (string, error) {
	query, err := query(str, t)
	if err != nil {
		log.Fatal("The query is empty")
	}
	req, err := http.NewRequest("POST", "http://localhost:4080/api/enron_mail/_search", strings.NewReader(query))
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(body), nil
}
