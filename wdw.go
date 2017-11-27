package main

import (
	"os"
	"log"
	"path/filepath"
	"strings"
	"net/http"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"gopkg.in/cheggaaa/pb.v1"
)

func main() {
	for _, url := range os.Args[1:] {
		log.Printf("MAIN: Try for %s", url)

		dirName := strings.Split(filepath.Base(url), ".")[0]
		log.Printf("MAIN: Create folder: %s", dirName)

		os.MkdirAll(dirName, os.ModePerm)

		links, err := findLinks(url)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		baseDir, _ := filepath.Abs("./")

		links = removeDuplicates(links)
		log.Printf("MAIN: Found %s links", len(links))
		for nm, link := range links {
			prefix := fmt.Sprintf("[%d / %d]", nm + 1, len(links))
			downloadFile(link, filepath.Join(baseDir, dirName)+"/", prefix)
		}
	}
	log.Println("* * *")
	log.Println("Done.")
}

func downloadFile(url string, dirToSave, prefix string) error {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	size := response.ContentLength
	bar := pb.New(int(size)).SetUnits(pb.U_BYTES)
	bar.Prefix(prefix)
	bar.Start()
	rd := bar.NewProxyReader(response.Body)

	fName := strings.Split(url, "/")

	file, err := os.Create(dirToSave + fName[len(fName)-1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if _, err := io.Copy(file, rd); err != nil {
		log.Fatal(err)
	}

	bar.Finish()
	return nil
}

func findLinks(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Panic(err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("findLinks: get %s: %s", url, resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("findLinks: parse %s as HTML: %v", url, err)
	}

	domain := strings.Split(url, "/")
	return visit(nil, doc, domain[0]+"//"+domain[1]+domain[2]), nil
}

func visit(links []string, n *html.Node, domain string) []string {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" && strings.Contains(a.Val, "webm") {
				links = append(links, domain+a.Val)
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		links = visit(links, c, domain)
	}
	return links
}

func removeDuplicates(elements []string) []string {
	encountered := map[string]bool{}
	var result []string
	for v := range elements {
		if encountered[elements[v]] == true {
		} else {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}
	return result
}
