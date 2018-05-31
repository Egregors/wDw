package main

import (
	"os"
	"fmt"
	"net/http"
	"strings"
	"golang.org/x/net/html"
	"io"
	"path/filepath"
	"sync"
)

// Download file
func fileDownloader(links <-chan string, path string, w *sync.WaitGroup) {
	defer w.Done()
	for url := range links {
		fmt.Println(url)

		response, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fileName := strings.Split(url, "/")
		file, err := os.Create(path + fileName[len(fileName)-1])
		if err != nil {
			fmt.Println(err)
			continue
		}

		if _, err := io.Copy(file, response.Body); err != nil {
			//fmt.Fprintln(b, err)
			fmt.Println(err)
			continue
		}

		response.Body.Close()
		file.Close()
	}
}

func findLinks(url string, links chan<- string) {
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	doc, _ := html.Parse(resp.Body)
	domain := strings.Split(url, "/")
	visit(links, doc, domain[0]+"//"+domain[1]+domain[2])

	defer close(links)
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

	for _, url := range os.Args[1:] {

		fmt.Println("URL: ", url)

		links := make(chan string)
		unseenLinks := make(chan string)
		w := &sync.WaitGroup{}

		dirName := strings.Split(filepath.Base(url), ".")[0]
		os.MkdirAll(dirName, os.ModePerm)
		baseDir, _ := filepath.Abs("./")
		dirToSave := filepath.Join(baseDir, dirName)

		for i := 0; i < queueSize; i++ {
			w.Add(1)
			go fileDownloader(unseenLinks, dirToSave+"/", w)
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

		w.Wait()
	}

	defer fmt.Println("DONE")
}
