package scraper

import (
	"fmt"
	"os"

	// "log"
	// import Colly
	"github.com/gocolly/colly"
)

type Dictionary map[string]string;

type TCEpisodeSpec struct {
	guests []string
	topics string
	segments []string
	continuity []string
	quotes []string
}

type TCMusicSpec struct {
	topFive []string
	songsPlayed []string
}

type TCContentSpec struct {
	url string
	name string
	episode TCEpisodeSpec
	music TCMusicSpec
}

/* OLD - See tc.go*/
func scraper(){
	args := os.Args
	url := args[1]
	collector := colly.NewCollector()

	/*
	Base Logic
	*/
	// whenever the collector is about to make a new request
    collector.OnRequest(func(r *colly.Request) {
        // print the url of that request
        fmt.Println("Visiting", r.URL)
    })
    collector.OnResponse(func(r *colly.Response) {
        fmt.Println("Got a response from", r.Request.URL)
    })
    collector.OnError(func(r *colly.Response, e error) {
        fmt.Println("Blimey, an error occurred!:", e)
    })


	/*
	Dig into pages
	*/
	collector.OnHTML(".mw-allpages-chunk a[href]", func(e *colly.HTMLElement) {
        link := e.Attr("href")
        text := e.Text
        fmt.Printf("Link found: %s -> %s\n", text, link)

		content := TCContentSpec{}

		// initialise a new recipe struct every time we visit a page
        // recipe := Recipe{}
        // initialise a new Dictionary object to stoer the ingredients mappings
        // ingredients_dictionary := Dictionary{}

        // assign the value of URL (the url we are visiting) to the recipe field
        content.url = url

		fmt.Printf("Content found: %s\n", content)

        // find the recipe title, assign it to the struct, and print it in the command line
        // content.name = main.ChildText(".mw-page-title-main")
        // println("Scraping recipe for:", content.name)
    })

    // Start scraping on a URL
    collector.Visit(url)
}