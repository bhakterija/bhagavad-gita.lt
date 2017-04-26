package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// BG is a top-level structure of the book
var BG Book

// defaultLangID - default to LT if no language specified
var defaultLangID = "lt"

var router *mux.Router

func main() {
	port := os.Getenv("PORT")

	if port == "" && len(os.Args) > 1 {
		port = os.Args[1]
	}

	if port == "" {
		log.Fatal("Port number must be specified either as a first argument or $PORT environment variable")
	}

	loadJSON()

	router = mux.NewRouter()
	router.HandleFunc("/", IndexHandler)
	router.HandleFunc("/{language:en|lt}", LangIndexHandler).Name("langIndex")

	router.HandleFunc("/{chapter:\\d{1,2}}", ChapterHandler)
	router.HandleFunc("/{language:en|lt}/{chapter:\\d{1,2}}", LangChapterHandler).Name("langChapter")

	router.HandleFunc("/{chapter:\\d{1,2}}/{verse:\\d{1,2}}", ChapterVerseHandler)
	router.HandleFunc("/{language:en|lt}/{chapter:\\d{1,2}}/{verse:\\d{1,2}}", LangChapterVerseHandler).Name("langChapterVerse")

	router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	//TODO: favicon, robots.txt
	router.HandleFunc("/favicon.ico", func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, "favicon.ico")
	})

	log.Fatal(http.ListenAndServe(":"+port, router))
}
