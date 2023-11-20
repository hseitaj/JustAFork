package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/temoto/robotstxt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// URLData holds information about each URL to be crawled.
type URLData struct {
	URL     string    // The URL to be crawled
	Created time.Time // Timestamp of URL creation or retrieval
	Links   []string
}
type YearData struct {
	Year string `json:"year"`
	Jan  string `json:"jan"`
	Feb  string `json:"feb"`
	Mar  string `json:"mar"`
	Apr  string `json:"apr"`
	May  string `json:"may"`
	Jun  string `json:"jun"`
	July string `json:"july"`
	Aug  string `json:"aug"`
	Sept string `json:"sept"`
	Oct  string `json:"oct"`
	Nov  string `json:"nov"`
	Dec  string `json:"dec"`
	Avg  string `json:"avg"`
}

type GasolineData struct {
	Year                     string `json:"year"`
	AverageGasolinePrices    string `json:"average_gasoline_prices"`
	AverageAnnualCPIForGas   string `json:"average_annual_cpi_for_gasoline"`
	GasPricesAdjustedForInfl string `json:"gas_prices_adjusted_for_inflation"`
}

type PropertyData struct {
	Status    string `json:"status"`
	Bedrooms  string `json:"bedrooms"`
	Bathrooms string `json:"bathrooms"`
	AcreLot   string `json:"acre_lot"`
	City      string `json:"city"`
	State     string `json:"state"`
	ZipCode   string `json:"zip_code"`
	HouseSize string `json:"house_size"`
	SoldDate  string `json:"prev_sold_date"`
	Price     string `json:"price"`
}

// crawlURL is responsible for crawling a single URL.
func crawlURL(urlData URLData, ch chan<- URLData, wg *sync.WaitGroup) {
	defer wg.Done() // Ensure the WaitGroup counter is decremented on function exit
	c := colly.NewCollector(
		colly.UserAgent(GetRandomUserAgent()), // Set a random user agent
	)
	// First, check if the URL is allowed by robots.txt rules
	allowed := isURLAllowedByRobotsTXT(urlData.URL)
	if !allowed {
		return // Skip crawling if not allowed
	}

	// Handler for errors during the crawl
	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Error occurred while crawling %s: %s\n", urlData.URL, err)
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		urlData.Links = append(urlData.Links, link)
	})

	// Handler for anchor tags found in HTML
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		fmt.Println("Found link:", link)
		// Here you can enqueue the link for further crawling or processing
	})

	// Handler for successful HTTP responses
	c.OnResponse(func(r *colly.Response) {
		if r.StatusCode == 200 {
			// Successful crawl, process the response here
			ch <- urlData // Send the URLData to the channel
			fmt.Printf("Crawled URL: %s\n", urlData.URL)
		} else {
			// Handle cases where the status code is not 200
			fmt.Printf("Non-200 status code while crawling %s: %d\n", urlData.URL, r.StatusCode)
		}
	})

	// Start the crawl
	c.Visit(urlData.URL)

	ch <- urlData
}

func createSiteMap(urls []URLData) error {
	siteMap := make(map[string][]string)
	for _, u := range urls {
		siteMap[u.URL] = u.Links
	}

	jsonData, err := json.Marshal(siteMap)
	err = ioutil.WriteFile("siteMap.json", jsonData, 0644)
	if err != nil {
		log.Printf("Error writing sitemap to file: %v\n", err)
		return err
	}

	log.Println("Sitemap created successfully.")
	return nil
}

// isURLAllowedByRobotsTXT checks if the given URL is allowed by the site's robots.txt.
func isURLAllowedByRobotsTXT(urlStr string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		log.Println("Error parsing URL:", err)
		return false
	}

	domain := parsedURL.Host
	robotsURL := "http://" + domain + "/robots.txt"

	resp, err := http.Get(robotsURL)
	if err != nil {
		log.Println("Error fetching robots.txt:", err)
		return true
	}

	data, err := robotstxt.FromResponse(resp)
	if err != nil {
		log.Println("Error parsing robots.txt:", err)
		return true
	}

	return data.TestAgent(urlStr, "GoEngine")
}

// threadedCrawl starts crawling the provided URLs concurrently.
func threadedCrawl(urls []URLData, concurrentCrawlers int) {
	var wg sync.WaitGroup
	ch := make(chan URLData, len(urls))

	rateLimitRule := &colly.LimitRule{
		DomainGlob:  "*",             // Apply to all domains
		Delay:       5 * time.Second, // Wait 5 seconds between requests
		RandomDelay: 5 * time.Second, // Add up to 5 seconds of random delay
	}

	log.Println("Starting crawling...")
	for _, urlData := range urls {
		wg.Add(1)

		go func(u URLData) {
			c := colly.NewCollector(
				colly.UserAgent(GetRandomUserAgent()),
			)
			c.Limit(rateLimitRule) // Set the rate limit rule

			crawlURL(u, ch, &wg)
		}(urlData)

		log.Println("Crawling URL:", urlData.URL)
		if len(urls) >= concurrentCrawlers {
			break
		}
	}

	log.Println("Waiting for crawlers to finish...")
	go func() {
		wg.Wait()
		close(ch)
		log.Println("All goroutines finished, channel closed.")
	}()

	var crawledURLs []URLData
	for urlData := range ch {
		crawledURLs = append(crawledURLs, urlData)
	}
	if err := createSiteMap(crawledURLs); err != nil {
		log.Println("Error creating sitemap:", err)
	}
}

// InitializeCrawling sets up and starts the crawling process.
func InitializeCrawling() {
	log.Println("Fetching URLs to crawl...")
	urlDataList := getURLsToCrawl()
	log.Println("URLs to crawl:", urlDataList)

	threadedCrawl(urlDataList, 10)
}

// getURLsToCrawl retrieves a list of URLs to be crawled.
func getURLsToCrawl() []URLData {

	return []URLData{
		{URL: "https://www.kaggle.com/search?q=housing+prices"},
		{URL: "http://books.toscrape.com/"},
		{URL: "https://www.kaggle.com/search?q=stocks"},
		{URL: "https://www.kaggle.com/search?q=stock+market"},
		{URL: "https://www.kaggle.com/search?q=real+estate"},
	}
}

func airdatatest() {
	urlll := "https://www.usinflationcalculator.com/inflation/airfare-inflation/"
	res, err := http.Get(urlll)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data []YearData

	doc.Find("table tbody tr").Each(func(rowIndex int, rowHtml *goquery.Selection) {
		if rowIndex == 0 {
			return
		}

		var yearData YearData
		rowHtml.Find("td").Each(func(cellIndex int, cellHtml *goquery.Selection) {
			switch cellIndex {
			case 0:
				yearData.Year = cellHtml.Text()
			case 1:
				yearData.Jan = cellHtml.Text()
			case 2:
				yearData.Feb = cellHtml.Text()
			case 3:
				yearData.Mar = cellHtml.Text()
			case 4:
				yearData.Apr = cellHtml.Text()
			case 5:
				yearData.May = cellHtml.Text()
			case 6:
				yearData.Jun = cellHtml.Text()
			case 7:
				yearData.July = cellHtml.Text()
			case 8:
				yearData.Aug = cellHtml.Text()
			case 9:
				yearData.Sept = cellHtml.Text()
			case 10:
				yearData.Oct = cellHtml.Text()
			case 11:
				yearData.Nov = cellHtml.Text()
			case 12:
				yearData.Dec = cellHtml.Text()
			case 13:
				yearData.Avg = cellHtml.Text()

			}
		})

		data = append(data, yearData)
	})

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("airfare_data.json", jsonData, 0644)
	if err != nil {
		log.Fatalf("Failed to write JSON data to file: %s", err)
	}

	log.Println("Airfare data written to airfare_data.json")
}

func scrapeInflationData() {
	urlll := "https://www.usinflationcalculator.com/inflation/current-inflation-rates/"
	res, err := http.Get(urlll)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data []YearData
	doc.Find("table tbody tr").Each(func(rowIndex int, rowHtml *goquery.Selection) {
		if rowIndex == 0 { // Skip the header row
			return
		}

		var yearData YearData
		rowHtml.Find("td").Each(func(cellIndex int, cellHtml *goquery.Selection) {
			text := cellHtml.Text()
			switch cellIndex {
			case 0:
				yearData.Year = cellHtml.Text()
			case 1:
				yearData.Jan = cellHtml.Text()
			case 2:
				yearData.Feb = cellHtml.Text()
			case 3:
				yearData.Mar = cellHtml.Text()
			case 4:
				yearData.Apr = cellHtml.Text()
			case 5:
				yearData.May = cellHtml.Text()
			case 6:
				yearData.Jun = cellHtml.Text()
			case 7:
				yearData.July = cellHtml.Text()
			case 8:
				yearData.Aug = cellHtml.Text()
			case 9:
				yearData.Sept = cellHtml.Text()
			case 10:
				yearData.Oct = cellHtml.Text()
			case 11:
				yearData.Nov = cellHtml.Text()
			case 12:
				yearData.Dec = cellHtml.Text()
			case 13:
				yearData.Avg = cellHtml.Text()
				yearData.Avg = text
			}
		})
		data = append(data, yearData)
	})

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("inflation_data.json", jsonData, 0644)
	if err != nil {
		log.Fatalf("Failed to write JSON data to file: %s", err)
	}

	fmt.Println("Inflation data written to inflation_data.json")
}

func scrapeGasInflationData() {
	urlll := "https://www.usinflationcalculator.com/gasoline-prices-adjusted-for-inflation/"
	res, err := http.Get(urlll)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data []GasolineData
	doc.Find("table tbody tr").Each(func(rowIndex int, rowHtml *goquery.Selection) {
		if rowIndex == 0 { // Skip the header row
			return
		}

		var gasData GasolineData
		rowHtml.Find("td").Each(func(cellIndex int, cellHtml *goquery.Selection) {
			text := cellHtml.Text()
			switch cellIndex {
			case 0:
				gasData.Year = text
			case 1:
				gasData.AverageGasolinePrices = text
			case 2:
				gasData.AverageAnnualCPIForGas = text
			case 3:
				gasData.GasPricesAdjustedForInfl = text
			}
		})
		data = append(data, gasData)
	})

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("gasoline_data.json", jsonData, 0644)
	if err != nil {
		log.Fatalf("Failed to write JSON data to file: %s", err)
	}

	fmt.Println("Gasoline data written to gasoline_data.json")
}

func scrapeHousingData() {
	urlll := "https://www.kaggle.com/datasets/ahmedshahriarsakib/usa-real-estate-dataset"
	res, err := http.Get(urlll)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var properties []PropertyData
	doc.Find(".sc-fLdTid.sc-eZkIzG.iXbLwD.cefCfQ").Each(func(i int, s *goquery.Selection) {
		var data PropertyData
		s.Find("div").Each(func(index int, item *goquery.Selection) {
			switch index {
			case 0:
				data.Status = item.Text()
			case 1:
				data.Bedrooms = item.Text()
			case 2:
				data.Bathrooms = item.Text()
			case 3:
				data.AcreLot = item.Text()
			case 4:
				data.City = item.Text()
			case 5:
				data.State = item.Text()
			case 6:
				data.ZipCode = item.Text()
			case 7:
				data.HouseSize = item.Text()
			case 8:
				data.SoldDate = item.Text()
			case 9:
				data.Price = item.Text()
			}
		})
		properties = append(properties, data)
	})

	jsonData, err := json.MarshalIndent(properties, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("property_data.json", jsonData, 0644)
	if err != nil {
		log.Fatalf("Failed to write JSON data to file: %s", err)
	}

	fmt.Println("Property data written to property_data.json")
}

func main() {
	InitializeCrawling()
	airdatatest()
	scrapeInflationData()
	scrapeGasInflationData()
	scrapeHousingData()
}
