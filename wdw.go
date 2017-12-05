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
	"io/ioutil"
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

		links = removeDuplicates(links)
		baseDir, _ := filepath.Abs("./")
		dirToSave := filepath.Join(baseDir, dirName)

		alreadyDownloaded := make(map[string]bool)
		files, err := ioutil.ReadDir(dirToSave)
		for _, f := range files {
			alreadyDownloaded[f.Name()] = true
		}

		log.Printf("MAIN: Found %d links", len(links))
		for nm, link := range links {
			prefix := fmt.Sprintf("[%d / %d]", nm+1, len(links))
			fName := strings.Split(link, "/")
			if !alreadyDownloaded[fName[len(fName)-1]] {
				downloadFile(link, dirToSave+"/", prefix)
			} else {
				log.Printf("%s already saved, next..", fName[len(fName)-1])
			}
		}
	}
	log.Println("* * *")
	log.Println("Done.")
}

func downloadFile(url string, dirToSave, prefix string) error {
	response, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return err
	}
	defer response.Body.Close()

	fName := strings.Split(url, "/")
	size := response.ContentLength
	bar := pb.New(int(size)).SetUnits(pb.U_BYTES)
	bar.Prefix(prefix)
	bar.Start()
	rd := bar.NewProxyReader(response.Body)

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
