package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Book struct {
	Id                      string
	Series                  string
	SeriesNo                string
	Title                   string
	Author                  string
	AuthorLf                string
	AdditionalAuthorrs      string
	Isbn                    string
	Isbn13                  string
	MyRating                string
	AvgRating               string
	Publisher               string
	Binding                 string
	Pages                   string
	PublishedYear           string
	OriginalPublicationYear string
	ReadDate                string
	AddDate                 string
	Bookshelves             string
	State                   string
}

func parseBookLine(record []string) Book {
	book := Book{}
	book.Id = record[0]
	title := record[1]
	if strings.Contains(title, "#") {
		// Handle series
		splits := strings.Split(title, "(")
		book.Title = strings.TrimSpace(splits[0])

		if len(splits) > 1 {
			seriesSplit := strings.Split(splits[1], "#")
			series := strings.TrimRight(strings.TrimSpace(seriesSplit[0]), ",")
			book.Series = series
			if len(seriesSplit) > 1 {
				book.SeriesNo = strings.TrimRight(strings.TrimSpace(seriesSplit[1]), ")")
			}
		}

	} else {
		book.Title = record[1]
	}

	book.Author = record[2]
	book.Isbn = strings.TrimLeft(strings.ReplaceAll(record[5], "\"", ""), "=")
	book.Isbn13 = strings.TrimLeft(strings.ReplaceAll(record[6], "\"", ""), "=")
	book.MyRating = record[7]
	book.Publisher = record[9]
	book.Binding = record[10]
	book.Pages = record[11]
	book.PublishedYear = record[12]
	book.OriginalPublicationYear = record[13]
	book.ReadDate = record[14]
	book.AddDate = record[15]
	book.Bookshelves = record[18]

	switch record[18] {
	case "read":
		book.State = "DONE"
	case "currently-reading":
		book.State = "INPROGRESS"
	case "to-read":
		book.State = "TODO"
	default:
		book.State = "TODO"

	}

	if record[18] == "read" {

	} else if record[18] == "to-read" {
		book.State = "TODO"
	}

	return book
}

func (v Book) String() string {
	if len(v.Series) > 0 {
		return fmt.Sprintf("Title: %s (Serie: %s/%s) by %s on shelve: %s\n", v.Title, v.Series, v.SeriesNo, v.Author, v.Bookshelves)
	} else {
		return fmt.Sprintf("Title: %s by %s on shelve: %s\n", v.Title, v.Author, v.Bookshelves)
	}
}

func writeString(input string, key string) string {
	out := ""
	if len(input) > 0 {
		out = fmt.Sprintf(":%s: %s\n", key, input)
	}

	return out
}

func (b Book) ToOrgMode() string {
	var buffer strings.Builder
	buffer.WriteString(fmt.Sprintf("** %s %s by %s\n", b.State, b.Title, b.Author))
	buffer.WriteString(":PROPERTIES:\n")
	buffer.WriteString(writeString(b.Title, "Title"))
	buffer.WriteString(writeString(b.Author, "Author"))
	buffer.WriteString(writeString(b.Series, "Series"))
	buffer.WriteString(writeString(b.SeriesNo, "Series#"))
	buffer.WriteString(writeString(b.Isbn13, "ISBN13"))
	buffer.WriteString(writeString(b.Isbn, "ISBN"))
	buffer.WriteString(writeString(b.Publisher, "Publisher"))
	buffer.WriteString(writeString(b.Pages, "Pages"))
	buffer.WriteString(writeString(b.OriginalPublicationYear, "FirstPublish"))
	buffer.WriteString(writeString(b.PublishedYear, "Published"))
	buffer.WriteString(writeString(b.AddDate, "Added"))
	buffer.WriteString(writeString(b.ReadDate, "Read"))

	buffer.WriteString(":END:")
	return buffer.String()
}

func main() {

	log.Println("Starting importer")
	if len(os.Args) < 2 {
		fmt.Println("Please specify file to read")
		os.Exit(1)
	}

	file := os.Args[1]
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("File reading error", err)
		os.Exit(1)
	}

	bookshelves := map[string][]Book{}

	//books := []Book{}
	r := csv.NewReader(bytes.NewReader(data))
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		book := parseBookLine(record)

		// Skip first line
		if book.Title != "Title" && book.Author != "Author" {
			bookshelves[book.Bookshelves] = append(bookshelves[book.Bookshelves], book)

		}

	}

	fmt.Println("#+TITLE: Books from Goodreads")
	fmt.Println("#+COMMENT: Imported by goodreads-to-org")
	fmt.Println("")

	shelves := 0
	noOfBooks := 0

	for shelf, books := range bookshelves {
		fmt.Println("* ", strings.ToUpper(shelf))
		shelves++
		noOfBooks += len(books)
		for _, book := range books {
			fmt.Println(book.ToOrgMode())
		}
	}
	log.Println("Parsing complete -", noOfBooks, "books on", shelves, "shelves")
}
