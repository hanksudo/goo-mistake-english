package main

import (
	"bytes"
	"crypto/tls"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"

	"os"

	"golang.org/x/net/html"
)

const baseURL = "https://dictionary.goo.ne.jp"

var contents []string

func extractArchiveList() []string {
	var urls []string

	archiveURL := baseURL + "/mistake_english/archive"
	log.Println("Start extracting archive list from...", archiveURL)

	// user http1
	http.DefaultClient.Transport = &http.Transport{
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}

	resp, err := http.Get(archiveURL)
	if err != nil {
		log.Printf("%s", err)
		os.Exit(1)
	}
	z := html.NewTokenizer(resp.Body)
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			return urls
		}

		t := z.Token()
		for _, a := range t.Attr {
			match, _ := regexp.MatchString(`/mistake_english/[\d+]`, a.Val)
			if match {
				urls = append(urls, a.Val)
			}
		}
	}
}
func f(n *html.Node) {
	if n.Data == "div" && n.Type == html.ElementNode {
		for _, a := range n.Attr {
			if a.Val == "content-box-english" {
				var buf bytes.Buffer
				_ = html.Render(&buf, n)
				contents = append(contents, buf.String())
				return
			}
		}
	}
	if n.FirstChild != nil {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
}

func extractContent(url string) {
	log.Println("Extract content from:", url)
	resp, _ := http.Get(baseURL + url)
	doc, _ := html.Parse(resp.Body)
	f(doc)
}

func main() {
	urls := extractArchiveList()
	log.Println("Count of pages:", len(urls))

	for _, url := range urls {
		extractContent(url)
	}

	log.Println("Rendering HTML...")
	f, _ := os.Create("index.html")
	tmpl, _ := template.ParseFiles("index.tmpl")
	tmpl.Execute(f, struct {
		Contents template.HTML
	}{Contents: template.HTML(strings.Join(contents, ""))})
	f.Close()

	log.Println("Done!")
}
