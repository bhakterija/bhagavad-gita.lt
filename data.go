package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
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
