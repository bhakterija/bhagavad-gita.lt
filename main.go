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
	Num         int
	Name        string
	Verses      []Verse
	PrevChapter int
	NextChapter int
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
	PrevVerse             [2]int // [ChapterNum, VerseNum] - Num, not Idx!
	NextVerse             [2]int
}

// BG is a top-level structure of the book
var BG Book

//TODO: title & keywords by language
var pageTitle = "Bhagavad-Gita As It Is. Translation and commentaries by HDG A.C.Bhaktivedanta Swami Prabhupada"
var pageKeywords = "Bhagavad Gita As It Is, Bhagavad-gītā, Bhagavad Gita, gita, ISKCON, Prabhupada, A.C. Bhaktivedanta Swami"

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

	// Set Prev and Next link values:
	for chapterIdx := range BG.Chapters {
		BG.Chapters[chapterIdx].PrevChapter = BG.Chapters[chapterIdx].Num - 1
		BG.Chapters[chapterIdx].NextChapter = BG.Chapters[chapterIdx].Num + 1
		if chapterIdx == 0 {
			BG.Chapters[chapterIdx].PrevChapter = 0
		} else if chapterIdx == len(BG.Chapters)-1 {
			BG.Chapters[chapterIdx].NextChapter = 0
		}
		for verseIdx, verse := range BG.Chapters[chapterIdx].Verses {

			// Default values for most of verses
			verse.PrevVerse[0] = chapterIdx + 1
			verse.PrevVerse[1] = verseIdx + 1 - 1
			verse.NextVerse[0] = chapterIdx + 1
			verse.NextVerse[1] = verseIdx + 1 + 1

			// Exceptions
			if chapterIdx == 0 && verseIdx == 0 {
				// Very first verse - no Prev
				verse.PrevVerse[0] = 0
				verse.PrevVerse[1] = 0
			} else if chapterIdx == len(BG.Chapters)-1 && verseIdx == len(BG.Chapters[chapterIdx].Verses)-1 {
				// Very last verse - no Next
				verse.NextVerse[0] = 0
				verse.NextVerse[1] = 0
			} else if verseIdx == 0 {
				// First verse of a chapter
				verse.PrevVerse[0]--
				verse.PrevVerse[1] = len(BG.Chapters[chapterIdx-1].Verses)
			} else if verseIdx == len(BG.Chapters[chapterIdx].Verses)-1 {
				// Last verse of a chapter
				verse.NextVerse[0]++
				verse.NextVerse[1] = 1
			}

			BG.Chapters[chapterIdx].Verses[verseIdx].PrevVerse = verse.PrevVerse
			BG.Chapters[chapterIdx].Verses[verseIdx].NextVerse = verse.NextVerse
		}
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
		//TODO: redirect
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

	if chapterNum < 1 || chapterNum > len(BG.Chapters) {
		fmt.Fprintf(w, "Chapter %v does not exist!\n", vars["chapter"])
		//TODO: redirect
	} else {
		chapter := BG.Chapters[chapterNum-1]
		// fmt.Fprintf(w, "[%s] %v. %s\n", vars["language"], chapterNum, BG.Chapters[chapterNum-1].Name)

		fp := path.Join("templates", "chapter.html")
		tmpl, err := template.ParseFiles(fp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create navigation arrows
		var prevHref, nextHref, upHref template.HTML
		if chapter.PrevChapter > 0 {
			prevURL, err := router.Get("langChapter").URL("language", vars["language"], "chapter", strconv.Itoa(chapter.PrevChapter))
			if err != nil {
				panic(err)
			}
			prevHref = template.HTML(fmt.Sprintf(`<a href="%s">&lt;&lt;</a>`, prevURL.String()))
		} else {
			prevHref = `&lt;&lt;`
		}

		if chapter.NextChapter > 0 {
			nextURL, err := router.Get("langChapter").URL("language", vars["language"], "chapter", strconv.Itoa(chapter.NextChapter))
			if err != nil {
				panic(err)
			}
			nextHref = template.HTML(fmt.Sprintf(`<a href="%s">&gt;&gt;</a>`, nextURL.String()))
		} else {
			nextHref = `&gt;&gt;`
		}

		//TODO: TOC route?
		// upURL, err := router.Get("langChapterVerse").URL("language", vars["language"], "chapter", strconv.Itoa(verse.NextVerse[0]), "verse", strconv.Itoa(verse.NextVerse[1]))
		// if err != nil {
		// 	panic(err)
		// }
		upHref = template.HTML(fmt.Sprintf(`<a href="%s">^</a>`, "/"))

		// construct verses list
		var versesList []template.HTML
		for _, verse := range chapter.Verses {
			verseURL, err := router.Get("langChapterVerse").URL("language", vars["language"], "chapter", strconv.Itoa(chapter.Num), "verse", strconv.Itoa(verse.Num))
			if err != nil {
				panic(err)
			}
			verseHref := template.HTML(fmt.Sprintf(`<a href="%s">%v.%v</a>`, verseURL.String(), chapter.Num, verse.Num))
			versesList = append(versesList, verseHref+" "+verse.Translation)
		}
		data := map[string]interface{}{
			"languageId":  vars["language"],
			"title":       pageTitle,
			"keywords":    pageKeywords,
			"chapterNum":  chapter.Num,
			"chapterName": chapter.Name,
			"verses":      versesList,
			"next":        nextHref,
			"prev":        prevHref,
			"up":          upHref,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// ChapterVerseHandler - handles route where chapter number and verse number are specified: /02/13/
func ChapterVerseHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chapterNum, _ := strconv.Atoi(vars["chapter"])
	verseNum, _ := strconv.Atoi(vars["verse"])

	if chapterNum == 0 || chapterNum > len(BG.Chapters) {
		fmt.Fprintf(w, "Chapter %v does not exist!\n", vars["chapter"])
		//TODO: redirect
	} else if verseNum == 0 || verseNum > len(BG.Chapters[chapterNum-1].Verses) {
		fmt.Fprintf(w, "Verse %v.%v does not exist!\n", vars["chapter"], vars["verse"])
		//TODO: redirect
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
		//TODO: redirect
	} else if verseNum == 0 || verseNum > len(BG.Chapters[chapterNum-1].Verses) {
		fmt.Fprintf(w, "Verse %v.%v does not exist!\n", vars["chapter"], vars["verse"])
		//TODO: redirect
	} else {
		fp := path.Join("templates", "verse.html")
		tmpl, err := template.ParseFiles(fp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		verse := BG.Chapters[chapterNum-1].Verses[verseNum-1]

		// fmt.Printf("Prev: %#v.%#v  Next: %#v.%#v \n", verse.PrevVerse[0], verse.PrevVerse[1], verse.NextVerse[0], verse.NextVerse[1])

		var prevHref, nextHref, upHref template.HTML
		if verse.PrevVerse[1] > 0 {
			prevURL, urlErr := router.Get("langChapterVerse").URL("language", vars["language"], "chapter", strconv.Itoa(verse.PrevVerse[0]), "verse", strconv.Itoa(verse.PrevVerse[1]))
			if urlErr != nil {
				panic(urlErr)
			}
			prevHref = template.HTML(fmt.Sprintf(`<a href="%s">&lt;&lt;</a>`, prevURL.String()))
		} else {
			prevHref = `&lt;&lt;`
		}

		if verse.NextVerse[1] > 0 {
			nextURL, urlErr := router.Get("langChapterVerse").URL("language", vars["language"], "chapter", strconv.Itoa(verse.NextVerse[0]), "verse", strconv.Itoa(verse.NextVerse[1]))
			if urlErr != nil {
				panic(urlErr)
			}
			nextHref = template.HTML(fmt.Sprintf(`<a href="%s">&gt;&gt;</a>`, nextURL.String()))
		} else {
			nextHref = `&gt;&gt;`
		}

		upURL, urlErr := router.Get("langChapter").URL("language", vars["language"], "chapter", strconv.Itoa(verse.NextVerse[0]))
		if urlErr != nil {
			panic(urlErr)
		}
		upHref = template.HTML(fmt.Sprintf(`<a href="%s">^</a>`, upURL.String()))

		// Join Synonyms string
		synonyms := ""
		separator := "; "
		for i, v := range BG.Chapters[chapterNum-1].Verses[verseNum-1].SynonymsSanskrit {
			if i == len(BG.Chapters[chapterNum-1].Verses[verseNum-1].SynonymsSanskrit)-1 { // last entry
				separator = "."
			}
			synonyms += v + "—" + BG.Chapters[chapterNum-1].Verses[verseNum-1].SynonymsTranslation[i] + separator
		}

		data := map[string]interface{}{
			"languageId": vars["language"],
			"title":      pageTitle,
			"keywords":   pageKeywords,
			"chapterNum": chapterNum,
			"verseNum":   verseNum,
			"synonyms":   synonyms,
			"verse":      BG.Chapters[chapterNum-1].Verses[verseNum-1],
			"next":       nextHref,
			"prev":       prevHref,
			"up":         upHref,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
