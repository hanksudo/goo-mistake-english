package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"os"

	"golang.org/x/net/html"
)

const baseURL = "http://dictionary.goo.ne.jp"

var contents []string

func extractArchiveList() []string {
	var urls []string
	resp, _ := http.Get(baseURL + "/mistake_english/archive")
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
	fmt.Println("Extract content from:", url)
	resp, _ := http.Get(baseURL + url)
	doc, _ := html.Parse(resp.Body)
	f(doc)
}

func main() {
	urls := extractArchiveList()
	fmt.Println("Count of pages:", len(urls))
	for _, url := range urls {
		extractContent(url)
	}

	fmt.Println("Rendering HTML...")
	f, _ := os.Create("index.html")
	tmpl, _ := template.ParseFiles("index.tmpl")
	tmpl.Execute(f, struct {
		Contents template.HTML
	}{Contents: template.HTML(strings.Join(contents, ""))})
	f.Close()
	fmt.Println("Done!")
}
