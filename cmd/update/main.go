package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/divan/num2words"
	"github.com/gocolly/colly/v2"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	outPath   = flag.String("out", "flags.json", "Path to save JSON data to")
	flagsPath = flag.String("flags", "images", "Path to save flag images to")
)

type Flag struct {
	Country     string   `json:"country"`
	Image       string   `json:"image"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

const target = "https://www.cia.gov/library/publications/resources/the-world-factbook/docs/flagsoftheworld.html"

func main() {
	var flags []*Flag

	c := colly.NewCollector()
	c.OnHTML(".wfb-modal-dialog", func(element *colly.HTMLElement) {
		src := element.ChildAttr(".modalFlagBox img", "src")
		if err := downloadFlag(src); err != nil {
			log.Fatalf("Unable to download flag: %v", err)
		}

		name := element.ChildText("span.countryName")
		desc := fixDescription(element.ChildText(".photogallery_captiontext"))

		flags = append(flags, &Flag{
			Country:     name,
			Description: desc,
			Image:       filepath.Join(*flagsPath, path.Base(src)),
			Keywords:    words(fmt.Sprintf("%s %s", name, desc)),
		})
	})
	if err := c.Visit(target); err != nil {
		log.Fatalf("Unable to retrieve flag data: %v", err)
	}

	f, err := os.Create(*outPath)
	if err != nil {
		log.Fatalf("Unable to create output file: %v", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(flags); err != nil {
		log.Fatalf("Unable to write output file: %v", err)
	}
}

var noteRegexp = regexp.MustCompile("note:.*$")

// fixDescription fixes some oddities found in descriptions
func fixDescription(text string) string {
	return strings.ReplaceAll(noteRegexp.ReplaceAllString(text, ""), " ", " ")
}

var unusedChars = regexp.MustCompile("[^a-zA-Z0-9]")

func words(text string) []string {
	found := make(map[string]bool)
	words := strings.Split(strings.ReplaceAll(strings.ToLower(text), "-", " "), " ")
	for i := range words {
		cleaned := strings.TrimSpace(unusedChars.ReplaceAllString(words[i], ""))
		if len(cleaned) == 0 {
			continue
		}

		if num, err := strconv.Atoi(cleaned); err == nil {
			if str := num2words.Convert(num); !strings.Contains(str, " ") {
				found[str] = true
				continue
			}
		}

		if len(cleaned) > 1 {
			found[cleaned] = true
		}
	}

	res := make([]string, 0, len(words))
	for i := range found {
		res = append(res, i)
	}
	return res
}

func downloadFlag(relativePath string) error {
	target := filepath.Join(*flagsPath, path.Base(relativePath))
	if _, err := os.Stat(target); err == nil {
		// File already exists, don't bother redownloading
		return nil
	}

	u, err := url.Parse(target)
	if err != nil {
		return err
	}
	u.Path = path.Join(path.Dir(u.Path), relativePath)

	resp, err := http.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath.Join(*flagsPath, path.Base(relativePath)))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
