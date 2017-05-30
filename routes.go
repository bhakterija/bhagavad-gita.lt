package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/mux"
)

//TODO: title & keywords by language
var pageTitle = "Bhagavad-Gita As It Is. Translation and commentaries by HDG A.C.Bhaktivedanta Swami Prabhupada"
var pageKeywords = "Bhagavad Gita As It Is, Bhagavad-gītā, Bhagavad Gita, gita, ISKCON, Prabhupada, A.C. Bhaktivedanta Swami"

// IndexHandler - default route "/" handler - redirect to /+defaultLangID
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	url, err := router.Get("langIndex").URL("language", defaultLangID)
	if err != nil {
		panic(err)
	}
	http.Redirect(w, r, url.String(), 301)
}

// LangIndexHandler - handles root+languageId: /lt/
func LangIndexHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fp := path.Join("templates", "toc.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// construct chapters list
	var chaptersList []template.HTML
	for _, chapter := range BG.Chapters {
		chapterURL, err := router.Get("langChapter").URL("language", vars["language"], "chapter", strconv.Itoa(chapter.Num))
		if err != nil {
			panic(err)
		}
		numHref := template.HTML(fmt.Sprintf(`<a href="%s">%v</a>`, chapterURL.String(), chapter.Num))
		nameHref := template.HTML(fmt.Sprintf(`<a href="%s">%v</a>`, chapterURL.String(), chapter.Name))
		chaptersList = append(chaptersList, "<td>"+numHref+"</td><td>"+nameHref+"</td>")
	}

	data := map[string]interface{}{
		"languageId": vars["language"],
		"title":      pageTitle,
		"keywords":   pageKeywords,
		"chapters":   chaptersList,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ChapterHandler - handles route where only chapter number is specified: /18/
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
			prevURL, prevErr := router.Get("langChapter").URL("language", vars["language"], "chapter", strconv.Itoa(chapter.PrevChapter))
			if prevErr != nil {
				panic(err)
			}
			prevHref = template.HTML(fmt.Sprintf(`<a href="%s">&lt;&lt;</a>`, prevURL.String()))
		} else {
			prevHref = `&lt;&lt;`
		}

		if chapter.NextChapter > 0 {
			nextURL, nextErr := router.Get("langChapter").URL("language", vars["language"], "chapter", strconv.Itoa(chapter.NextChapter))
			if nextErr != nil {
				panic(err)
			}
			nextHref = template.HTML(fmt.Sprintf(`<a href="%s">&gt;&gt;</a>`, nextURL.String()))
		} else {
			nextHref = `&gt;&gt;`
		}

		upURL, upErr := router.Get("langIndex").URL("language", vars["language"])
		if upErr != nil {
			panic(err)
		}
		upHref = template.HTML(fmt.Sprintf(`<a href="%s">^</a>`, upURL.String()))

		// construct verses list
		var versesList []template.HTML
		for _, verse := range chapter.Verses {
			verseURL, err := router.Get("langChapterVerse").URL("language", vars["language"], "chapter", strconv.Itoa(chapter.Num), "verse", strconv.Itoa(verse.Num))
			if err != nil {
				panic(err)
			}
			verseNumHref := template.HTML(fmt.Sprintf(`<a href="%s">%v.%v</a>`, verseURL.String(), chapter.Num, verse.Num))
			// verseIASTHref := template.HTML(fmt.Sprintf(`<a href="%s">%v</a>`, verseURL.String(), strings.Join(verse.IAST, " | ")))
			verseTranslationHref := template.HTML(fmt.Sprintf(`<a href="%s">%v</a>`, verseURL.String(), verse.Translation))
			// versesList = append(versesList, "<td rowspan=\"2\" valign=\"top\">"+verseNumHref+"</td><td>"+verseIASTHref+"</td></tr> <tr><td>"+verseTranslationHref+"</td></tr>")
			versesList = append(versesList, "<td valign=\"top\">"+verseNumHref+"</td><td>"+verseTranslationHref+"</td></tr>")
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
		var synonyms template.HTML
		var separator template.HTML = "; "
		for i, v := range BG.Chapters[chapterNum-1].Verses[verseNum-1].SynonymsSanskrit {
			if i == len(BG.Chapters[chapterNum-1].Verses[verseNum-1].SynonymsSanskrit)-1 { // last entry
				separator = "."
			}
			synonyms += template.HTML(v) + "—" + BG.Chapters[chapterNum-1].Verses[verseNum-1].SynonymsTranslation[i] + separator
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
