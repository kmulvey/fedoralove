package main

import (
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.InfoLevel)

	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("github.com"),
	)

	// On every a element which has href attribute call callback
	c.OnHTML("body", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "sudo apt") && !(strings.Contains(e.Text, "sudo dnf ") || !strings.Contains(e.Text, "sudo yum ")) {
			log.Infof("found apt at: %s", e.Request.URL)
		}
	})

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		var link = e.Attr("href")
		var visit, err = interestingLink(link)
		if err != nil {
			log.Errorf("error parsing link: %s, error: %w", link, err)
		}
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		if visit {
			time.Sleep(time.Second) // dont ddos github
			c.Visit(e.Request.AbsoluteURL(link))
		}
	})

	// Before making a request print "Visiting ..."
	var numVisited int
	c.OnRequest(func(r *colly.Request) {
		numVisited++
		if numVisited%100 == 0 {
			log.Infof("visited: %d", numVisited)
		}
		log.Tracef("Visiting: %s", r.URL.String())
	})

	// seed link
	c.Visit("https://github.com/sindresorhus/awesome")
}

func interestingLink(link string) (bool, error) {
	var u, err = url.Parse(link)
	if err != nil {
		return false, err
	}
	var path = u.EscapedPath()
	var pathArr = strings.Split(path, "/")

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
