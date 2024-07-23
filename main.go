package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"log"

	// "strings"

	"github.com/gocolly/colly"
	// "github.com/supabase-community/postgrest-go"

	"github.com/PuerkitoBio/goquery"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Episode struct {
	Url     				string 			`bson:"url,omitempty"`
	Title					string			`bson:"title,omitempty"`
	EpisodeNo 				string			`bson:"episode_no,omitempty"`
	Date					string			`bson:"date,omitempty"`
	Guests 					[]string		`bson:"guests,omitempty"`
	Top5ComparisonYear 		string			`bson:"top_5_comparison_year,omitempty"`
	Notes  					string			`bson:"notes,omitempty"`

}

func main() {
	// 
	// Connect to Supabase / Postgres
	// 
	// projectURL := os.Getenv("SUPABASE_PROJECT_URL")
	// apiKey := os.Getenv("SUPABASE_API_KEY")

	// if projectURL == "" || apiKey == "" {
	// 	log.Fatal("Supabase URL or API key not set in environment variables")
	// }

	// // Initialize Postgrest client
	// client, err := supabase.NewClient(projectURL, apiKey, nil)
	// if err != nil {
    // 	fmt.Println("cannot initalize client", err)
  	// }
	username := os.Getenv("MONGO_ATLAS_USERNAME")
    password := os.Getenv("MONGO_ATLAS_PASSWORD")
	// mongodb+srv://rudypenajr:<password>@cluster0.6ay7bhy.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0
	mongoURI := fmt.Sprintf("mongodb+srv://%s:%s@cluster0.6ay7bhy.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0", username, password)
	fmt.Printf(`values: %s, %s \n %s`, username, password, mongoURI)
    clientOptions := options.Client().ApplyURI(mongoURI)

    // Connect to MongoDB
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }
    defer func() {
        if err = client.Disconnect(context.TODO()); err != nil {
            log.Fatal(err)
        }
    }()

    // Check the connection
    err = client.Ping(context.TODO(), nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Connected to MongoDB!")

	// Get a handle for your collection
    collection := client.Database("tc-webscraper").Collection("episodes")


	// 
	// Initialize Colly Collector
	// 
	c := colly.NewCollector(
		colly.AllowedDomains("the-time-crisis-universe.fandom.com"),
	)

	var episodes []Episode

	// Visit main page to get all links
	c.OnHTML(".article-table", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(i int, row *colly.HTMLElement) {
            if i == 0 {
                // Skip header row
                return
            }
            
            // Extract data from the row
            episodeNo := row.ChildText("td:nth-child(1)")
			title := row.ChildText("td:nth-child(2)")
			url := `https://the-time-crisis-universe.fandom.com/` + row.ChildText("td:nth-child(2) a[href]")
            date := row.ChildText("td:nth-child(3)")

			top5ComparisonYear := row.ChildText("td:nth-child(5)")
			notes := row.ChildText("td:nth-child(6)")
            

            // Create a new episode struct and add it to the slice
            episode := Episode{
                Title:       title,
				Url: url,
				EpisodeNo: episodeNo,
				Date: date,
				// Guests: guests,
				Top5ComparisonYear: top5ComparisonYear,
				Notes: notes,
            }
            episodes = append(episodes, episode)
        })
	})

	// 
	// Start scraping
	// 
	c.Visit("https://the-time-crisis-universe.fandom.com/wiki/Episode_Guide");
	// c.Visit("https://the-time-crisis-universe.fandom.com/wiki/Special:AllPages")

	//
	// Update Guests
	//
	updateGuests(&episodes)
	 
	// Convert []Episode to []interface{}
    var interfaceSlice []interface{}
    for _, e := range episodes {
        interfaceSlice = append(interfaceSlice, e)
    }

    insertResult, err := collection.InsertMany(context.TODO(), interfaceSlice)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Inserted documents with IDs: %v\n", insertResult.InsertedIDs)
	


	// 
	// Convert `links`` to JSON
	// 
	// jsonData, err := json.MarshalIndent(episodes, "", " ")
	// if err != nil {
	// 	log.Fatalf("Failed to Marshal JSON: %v", err)
	// }
	// fmt.Println(string(jsonData))

	// response, body, err := client.From("episodes").Insert(jsonData, false, "", "", "").
	// response, body, err := client.From("episodes").Insert(&jsonData, true, "", "", "").Execute()
	// if err != nil {
	// 	log.Printf("Failed to insert data: %v\n", err)
	// 	log.Printf("Response: %v\nBody: %s\n", response, string(body))
	// 	return
	// }
	// log.Printf("Data inserted successfully! Response: %v\nBody: %s\n", response, string(body))
	// log.Printf("Data inserted successfully! Response: %v\n", response)

	// Get a handle for your collection
    // Insert a single document
    // insertResult, err := collection.InsertMany(context.TODO(), jsonData)
    // if err != nil {
        // log.Fatal(err)
    // }

    // fmt.Printf("Inserted document with ID: %v\n", insertResult.InsertedID)

}

func updateGuests(episodes *[]Episode) {
	// "https://the-time-crisis-universe.fandom.com/wiki/Episode_Guide"
	// Request the HTML page.
  	res, err := http.Get("https://the-time-crisis-universe.fandom.com/wiki/Episode_Guide")
  	if err != nil {
    	log.Fatalf("Failed to fetch page using goquery: %v", err)
  	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("Status Code Error using goquery: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
  	doc, err := goquery.NewDocumentFromReader(res.Body)
  	if err != nil {
		log.Fatalf("Loading HTML document using goquery: %v", err)
  	}

	// Iterate over Tables
	var counter = 0
	doc.Find(".article-table").Each(func(tableIdx int, t *goquery.Selection) {
		// Iterate over Rows in Tables
		t.Find("tr").Each(func(rowIdx int, r *goquery.Selection) {
			if rowIdx == 0 {
				return
			}
			
			td, err := r.Find("td:nth-child(4)").Html()
			if err != nil {
				log.Fatalf("Error parsing Guests using goquery: %v", err)
			}

			var guests []string
			if td == "—" {
				(*episodes)[counter].Guests = guests
				return
			}

			r.Find("td:nth-child(4)").Contents().Each(func(_ int, s *goquery.Selection) {
				// Handle <a> tags
				if s.Is("a") {
					// fmt.Printf("Anchor %d: %v\n", i, s.Text())
					guests = append(guests, s.Text())
					return;
				}

				if s.Is("span") {
					// fmt.Printf("Anchor %d: %v\n", i, s.Text())
					guests = append(guests, s.Text())
					return;
				}

				// Handle text nodes (e.g., text between <br> tags)
				if goquery.NodeName(s) == "#text" {
					text := strings.TrimSpace(s.Text())
					if text != "" {
						// fmt.Printf("Text %d: %v\n", i, text)
						guests = append(guests, text)
					}
				}
			})

			(*episodes)[counter].Guests = guests
			counter += 1
		})
	})
}