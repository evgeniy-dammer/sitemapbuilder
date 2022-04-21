package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	link "github.com/evgeniy-dammer/htmllinkparser"
)

const xmlns = "http://sitemaps.org/schemas/sitemap/0.9"

type loc struct {
	Value string `xml:"loc"`
}

type urlSet struct {
	Urls  []loc  `xml:"url"`
	Xmlns string `xml:"xmlns,attr"`
}

func bfs(urlStr string, maxDepth int) []string {

	seen := make(map[string]struct{})

	var q map[string]struct{}

	nq := map[string]struct{}{
		urlStr: {},
	}

	for i := 0; i <= maxDepth; i++ {
		q, nq = nq, make(map[string]struct{})

		if len(q) == 0 {
			break
		}

		for url := range q {
			if _, ok := seen[url]; ok {
				continue
			}
			seen[url] = struct{}{}

			for _, link := range getPages(url) {
				nq[link] = struct{}{}
			}
		}
	}

	result := make([]string, 0, len(seen))

	for url := range seen {
		result = append(result, url)
	}

	return result
}

func withPrefix(pfx string) func(string) bool {
	return func(link string) bool {
		return strings.HasPrefix(link, pfx)
	}
}

func filterBaseLinks(links []string, keepFn func(string) bool) []string {
	var baseLinks []string

	for _, link := range links {
		if keepFn(link) {
			baseLinks = append(baseLinks, link)
		}
	}

	return baseLinks
}

func getAllLinksOnPage(body io.Reader, base string) []string {
	var allLinks []string

	links, _ := link.Parse(body)

	for _, l := range links {
		switch {
		case strings.HasPrefix(l.Href, "/"):
			allLinks = append(allLinks, base+l.Href)
		case strings.HasPrefix(l.Href, "http"):
			allLinks = append(allLinks, l.Href)
		}
	}

	return allLinks
}

func getPages(urlFlag string) []string {
	responce, err := http.Get(urlFlag)

	if err != nil {
		return []string{}
	}

	defer responce.Body.Close()

	reqUrl := responce.Request.URL

	baseUrl := &url.URL{
		Scheme: reqUrl.Scheme,
		Host:   reqUrl.Host,
	}

	base := baseUrl.String()

	return filterBaseLinks(getAllLinksOnPage(responce.Body, base), withPrefix(base))
}

func main() {
	urlFlag := flag.String("url", "https://gophercises.com", "Url tha you want to build sitemap for")
	maxDepth := flag.Int("depth", 3, "Maximum number of links deep to traverse")
	flag.Parse()

	pages := bfs(*urlFlag, *maxDepth)

	toXml := urlSet{
		Xmlns: xmlns,
	}

	for _, page := range pages {
		toXml.Urls = append(toXml.Urls, loc{page})
	}

	fmt.Print(xml.Header)

	encoder := xml.NewEncoder(os.Stdout)
	encoder.Indent("", "	")
	if err := encoder.Encode(toXml); err != nil {
		panic(err)
	}
	fmt.Println()
}
