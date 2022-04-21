package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	link "github.com/evgeniy-dammer/htmllinkparser"
)

func main() {
	urlFlag := flag.String("url", "https://gophercises.com", "Url tha you want to build sitemap for")

	flag.Parse()

	responce, err := http.Get(*urlFlag)

	if err != nil {
		panic(err)
	}

	defer responce.Body.Close()

	reqUrl := responce.Request.URL

	baseUrl := &url.URL{
		Scheme: reqUrl.Scheme,
		Host:   reqUrl.Host,
	}

	base := baseUrl.String()

	links, err := link.Parse(responce.Body)

	if err != nil {
		panic(err)
	}

	var hrefs []string

	for _, l := range links {
		switch {
		case strings.HasPrefix(l.Href, "/"):
			hrefs = append(hrefs, base+l.Href)
		case strings.HasPrefix(l.Href, "http"):
			hrefs = append(hrefs, l.Href)
		}
	}

	for _, href := range hrefs {
		fmt.Println(href)
	}
}
