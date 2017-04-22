package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/gorilla/mux"
)

// Book - top level structure of a book
type Book struct {
	Preface      string
	Introduction string
	Chapters     [18]Chapter
}

// Chapter - individual chapter in a book
type Chapter struct {
	Num    int
	Name   string
	Verses []Verse
}

// Verse - individual verse in a chapter
type Verse struct {
	Num                   int
	Devanagari            []string
	DevanagariWordTimings []float32
	IAST                  []string
	IASTWordTimings       []float32
	SynonymsSanskrit      []string
	SynonymsTranslation   []string
	Translation           template.HTML
	Purport               []template.HTML
}

// BG is a top-level structure of the book
var BG Book

// defaultLangID - default to LT if no language specified
var defaultLangID = "lt"

var router *mux.Router

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = os.Args[1]
	}

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	loadJSON()

	router = mux.NewRouter()
	router.HandleFunc("/", IndexHandler)
	// router.HandleFunc("/{language:en|lt}", LangIndexHandler).Name("langIndexHandler")

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

func loadJSON() {
	data, err := ioutil.ReadFile("./public/texts/lt/83.json")
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(2)
	}
	if err := json.Unmarshal(data, &BG); err != nil {
		log.Fatalf("JSON unmarshalling failed: %s", err)
	}
}

// IndexHandler - default route "/" handler
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome!\n")
}

// ChapterHandler - handles route where only chapter number is specified: /01/
func ChapterHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chapterNum, _ := strconv.Atoi(vars["chapter"])

	if chapterNum == 0 || chapterNum > len(BG.Chapters) {
		fmt.Fprintf(w, "Chapter %v does not exist!\n", vars["chapter"])
	} else {
		url, err := router.Get("langChapter").URL("language", defaultLangID, "chapter", vars["chapter"])
		if err != nil {
			panic(err)
		}
		//fmt.Fprintf(w, "%v. %s\n", chapterNum, BG.Chapters[chapterNum-1].Name)
		http.Redirect(w, r, url.String(), 301)
	}
}

// LangChapterHandler - handles route where language and chapter are specified: /lt/01/
func LangChapterHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	//fmt.Fprintf(w, "LangChapterHandler: %v<br>", vars)
	chapterNum, _ := strconv.Atoi(vars["chapter"])
	//langId := vars["language"]

	if chapterNum == 0 || chapterNum > len(BG.Chapters) {
		fmt.Fprintf(w, "Chapter %v does not exist!\n", vars["chapter"])
	} else {
		fmt.Fprintf(w, "[%s] %v. %s\n", vars["language"], chapterNum, BG.Chapters[chapterNum-1].Name)
	}
}

// ChapterVerseHandler - handles route where chapter number and verse number are specified: /02/13/
func ChapterVerseHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chapterNum, _ := strconv.Atoi(vars["chapter"])
	verseNum, _ := strconv.Atoi(vars["verse"])

	if chapterNum == 0 || chapterNum > len(BG.Chapters) {
		fmt.Fprintf(w, "Chapter %v does not exist!\n", vars["chapter"])
	} else if verseNum == 0 || verseNum > len(BG.Chapters[chapterNum-1].Verses) {
		fmt.Fprintf(w, "Verse %v.%v does not exist!\n", vars["chapter"], vars["verse"])
	} else {
		url, err := router.Get("langChapterVerse").URL("language", defaultLangID, "chapter", vars["chapter"], "verse", vars["verse"])
		if err != nil {
			panic(err)
		}
		//fmt.Fprintf(w, "%v. %s\n", chapterNum, BG.Chapters[chapterNum-1].Name)
		http.Redirect(w, r, url.String(), 301)
	}
}

// LangChapterVerseHandler - handles route where language, chapter number and verse number are specified: /lt/02/13/
func LangChapterVerseHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chapterNum, _ := strconv.Atoi(vars["chapter"])
	verseNum, _ := strconv.Atoi(vars["verse"])

	if chapterNum == 0 || chapterNum > len(BG.Chapters) {
		fmt.Fprintf(w, "Chapter %v does not exist!\n", vars["chapter"])
	} else if verseNum == 0 || verseNum > len(BG.Chapters[chapterNum-1].Verses) {
		fmt.Fprintf(w, "Verse %v.%v does not exist!\n", vars["chapter"], vars["verse"])
	} else {

		fp := path.Join("templates", "verse.html")
		tmpl, err := template.ParseFiles(fp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		prevChapterNum := chapterNum
		prevVerseNum := verseNum - 1
		if chapterNum == 1 && verseNum == 1 { // already first chapter - no previous
			prevChapterNum = 0
			prevVerseNum = 0
		} else {
			if prevVerseNum < 1 {
				prevChapterNum--
				prevVerseNum = len(BG.Chapters[prevChapterNum-1].Verses)
			}
		}

		nextChapterNum := chapterNum
		nextVerseNum := verseNum + 1
		if nextVerseNum > len(BG.Chapters[chapterNum-1].Verses) {
			nextVerseNum = 1
			nextChapterNum++
		}
		if nextChapterNum > len(BG.Chapters) { // already last chapter - no next
			nextVerseNum = 0
			nextChapterNum = 0
		}

		prevURL, err := router.Get("langChapterVerse").URL("language", vars["language"], "chapter", strconv.Itoa(prevChapterNum), "verse", strconv.Itoa(prevVerseNum))
		if err != nil {
			panic(err)
		}
		nextURL, err := router.Get("langChapterVerse").URL("language", vars["language"], "chapter", strconv.Itoa(nextChapterNum), "verse", strconv.Itoa(nextVerseNum))
		if err != nil {
			panic(err)
		}

		var prevHref, nextHref template.HTML
		if prevChapterNum > 0 {
			prevHref = template.HTML(fmt.Sprintf(`<a href="%s">&lt;&lt;</a>`, prevURL.String()))
		} else {
			prevHref = `&lt;&lt;`
		}

		if nextChapterNum > 0 {
			nextHref = template.HTML(fmt.Sprintf(`<a href="%s">&gt;&gt;</a>`, nextURL.String()))
		} else {
			nextHref = `&gt;&gt;`
		}

		// Make Synonyms section
		synonyms := ""
		separator := "; "
		for i, v := range BG.Chapters[chapterNum-1].Verses[verseNum-1].SynonymsSanskrit {
			if i == len(BG.Chapters[chapterNum-1].Verses[verseNum-1].SynonymsSanskrit)-1 { // last entry
				separator = "."
			}
			synonyms += v + "—" + BG.Chapters[chapterNum-1].Verses[verseNum-1].SynonymsTranslation[i] + separator
		}

		//TODO: title & keywords by language
		data := map[string]interface{}{
			"languageId": vars["language"],
			"title":      "Bhagavad-Gita As It Is. Translation and commentaries by HDG A.C.Bhaktivedanta Swami Prabhupada",
			"keywords":   "Bhagavad Gita As It Is, Bhagavad-gītā, Bhagavad Gita, gita, ISKCON, Prabhupada, A.C. Bhaktivedanta Swami",
			"chapterNum": chapterNum,
			"verseNum":   verseNum,
			"synonyms":   synonyms,
			"verse":      BG.Chapters[chapterNum-1].Verses[verseNum-1],
			"next":       nextHref,
			"prev":       prevHref,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
