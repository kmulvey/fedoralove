package main

import (
	"bufio"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
)

const staticBool = false

//const startURL = "https://github.com/kmulvey/fedoralove"
const startURL = "https://github.com/sindresorhus/awesome"

var reservedPaths = map[string]bool{
	"about":            staticBool,
	"customer-stories": staticBool,
	"enterprise":       staticBool,
	"features":         staticBool,
	"fluidicon.png":    staticBool,
	"login":            staticBool,
	"notifications":    staticBool,
	"pricing":          staticBool,
	"readme":           staticBool,
	"security":         staticBool,
	"signup":           staticBool,
	"site":             staticBool,
	"team":             staticBool,
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetLevel(log.InfoLevel)

	// open the skip file
	log.Info("reading log file")
	var processedLog, err = os.OpenFile("scraped.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Fatalf("error opening scraped file: %s", err.Error())
	}
	defer func() {
		var err = processedLog.Close()
		if err != nil {
			log.Fatalf("error closing scraped file: %s", err.Error())
		}
	}()

	var skipMap = getSkipMap(processedLog)

	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("github.com"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Safari/537.36"),
		colly.IgnoreRobotsTxt(),
	)

	// On every a element which has href attribute call callback
	c.OnHTML("body", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "sudo apt") && (!strings.Contains(e.Text, "sudo dnf ") || !strings.Contains(e.Text, "sudo yum ")) {
			log.Infof("found apt at: %s", e.Request.URL)
		}
	})

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		var link = e.Attr("href")
		var visit, err = interestingLink(link)
		if err != nil {
			log.Errorf("error parsing link: %s, error: %s", link, err.Error())
		}
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		var _, alreadyVisited = skipMap[e.Request.AbsoluteURL(link)]
		if visit && !alreadyVisited {
			time.Sleep(time.Millisecond * 500) // dont ddos github
			c.Visit(e.Request.AbsoluteURL(link))
		}
	})

	// Before making a request print "Visiting ..."
	var numVisited int
	c.OnRequest(func(r *colly.Request) {
		numVisited++
		if numVisited%10 == 0 {
			log.Infof("visited: %d", numVisited)
		}
		log.Tracef("Visiting: %s", r.URL.String())
	})

	// put url in scraped file so we dont have to scrape it again
	c.OnResponse(func(r *colly.Response) {
		if r.Request.URL.String() != startURL {
			var _, err = processedLog.WriteString(r.Request.URL.String() + "\n")
			if err != nil {
				log.Fatalf("error writing to scraped file: %s", err.Error())
			}
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Errorf("Something went wrong: status: %d, error: %s", r.StatusCode, err.Error())
		if r.StatusCode == http.StatusTooManyRequests {
			time.Sleep(time.Minute * 5)
			log.Info("got StatusTooManyRequests response, sleeping for 5 minutes")
		}
	})

	// seed link
	c.Visit(startURL)
}

func interestingLink(link string) (bool, error) {
	var u, err = url.Parse(link)
	if err != nil {
		return false, err
	}
	var path = u.EscapedPath()
	var pathArr = strings.Split(path, "/")
	if len(pathArr) < 3 {
		return false, nil
	}

	// skip non repo pages
	if _, exists := reservedPaths[pathArr[1]]; exists {
		return false, nil
	}

	// repo homepage
	if len(pathArr) == 3 {
		return true, nil
	}

	// README.md, we only want the latest version so we only look at master or main branches
	if len(pathArr) == 6 {
		if pathArr[len(pathArr)-1] == "README.md" && (pathArr[len(pathArr)-2] == "master" || pathArr[len(pathArr)-2] == "main") && pathArr[len(pathArr)-3] == "blob" {
			return true, nil
		}
	}

	return false, nil
}

// getSkipMap read the log from the last time this was run and
// puts those filenames in a map so we dont have to process them again
// If you want to reprocess, just delete the file
func getSkipMap(processedImages *os.File) map[string]bool {

	var scanner = bufio.NewScanner(processedImages)
	scanner.Split(bufio.ScanLines)
	var compressedFiles = make(map[string]bool)

	for scanner.Scan() {
		compressedFiles[scanner.Text()] = staticBool
	}

	return compressedFiles
}
