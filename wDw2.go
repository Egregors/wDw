package main

import (
	"os"
	"fmt"
	"net/http"
	"strings"
	"golang.org/x/net/html"
	"io"
	"bytes"
	"path/filepath"
)

func fileDownloader(links <-chan string, path string) {
	b := &bytes.Buffer{}
	defer fmt.Println(b)

	for url := range links {
		response, err := http.Get(url)
		if err != nil {
			fmt.Println(b, err)
		}
		defer response.Body.Close()

		fileName := strings.Split(url, "/")
		file, err := os.Create(path + fileName[len(fileName)-1])
		if err != nil {
			fmt.Println(b, err)
		}
		defer file.Close()

		if _, err := io.Copy(file, response.Body); err != nil {
			fmt.Println(b, err)
		}
		fmt.Println(b, "done: "+url)
	}
}

func findLinks(url string, links chan<- string) error {
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	doc, _ := html.Parse(resp.Body)
	domain := strings.Split(url, "/")
	visit(links, doc, domain[0]+"//"+domain[1]+domain[2])

	defer close(links)
	return nil
}

func visit(links chan<- string, n *html.Node, domain string) []string {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" && (strings.Contains(a.Val, "webm") || strings.Contains(a.Val, "mp4")) {
				links <- domain + a.Val
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		visit(links, c, domain)
	}
	return nil
}

func main() {
	const queueSize = 10

	links := make(chan string)
	unseenLinks := make(chan string)

	url := os.Args[1:2][0]
	fmt.Println("URL: ", url)

	dirName := strings.Split(filepath.Base(url), ".")[0]
	os.MkdirAll(dirName, os.ModePerm)
	baseDir, _ := filepath.Abs("./")
	dirToSave := filepath.Join(baseDir, dirName)

	for i := 0; i < queueSize; i++ {
		go fileDownloader(unseenLinks, dirToSave+"/")
	}

	go findLinks(url, links)

	seen := make(map[string]bool)
	for link := range links {
		if !seen[link] {
			seen[link] = true
			unseenLinks <- link
		}
	}
	close(unseenLinks)
}
