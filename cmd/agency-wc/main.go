package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kfelter/ecfr-miner/ecfr"
)

type AgencyResponse struct {
	Agencies []Agency `json:"agencies"`
}

type Agency struct {
	Name          string  `json:"name"`
	ShortName     string  `json:"short_name"`
	DisplayName   string  `json:"display_name"`
	SortableName  string  `json:"sortable_name"`
	Slug          string  `json:"slug"`
	Children      []Child `json:"children"`
	CfrReferences []Ref   `json:"cfr_references"`
}

type Child struct {
	Name          string `json:"name"`
	ShortName     string `json:"short_name"`
	DisplayName   string `json:"display_name"`
	SortableName  string `json:"sortable_name"`
	Slug          string `json:"slug"`
	CfrReferences []Ref  `json:"cfr_references"`
}

type Ref struct {
	Title    int    `json:"title"`
	Chapter  string `json:"chapter"`
	Subtitle string `json:"subtitle"`
	Part     string `json:"part"`
}

func main() {
	flag.Parse()

	type titleDir struct {
		path string
		year string
	}

	titlesDirs := []titleDir{}

	// look for all the directories that start with titles_
	// dir looks like titles_2025
	// collect all the dirs in an array with the path and year

	filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && strings.HasPrefix(d.Name(), "titles_") {
			timeStr := strings.TrimPrefix(d.Name(), "titles_")

			titlesDirs = append(titlesDirs, titleDir{path: path, year: timeStr})
			fmt.Println("Found directory", path)
		}
		return nil
	})

	// get the agency data from api
	// curl -X GET "https://www.ecfr.gov/api/admin/v1/agencies.json" -H "accept: application/json"
	res, err := http.Get("https://www.ecfr.gov/api/admin/v1/agencies.json")
	if err != nil {
		fmt.Println("Error getting agencies", err)
		log.Fatal(err)
	}

	defer res.Body.Close()
	// check if the response status is OK
	if res.StatusCode != http.StatusOK {
		fmt.Println("Error getting agencies", res.Status)
		log.Fatalf("Error: %s", res.Status)
	}

	// save the raw response body to a file
	out, err := os.Create("agencies.json")
	if err != nil {
		fmt.Println("Error creating file", err)
		log.Fatal(err)
	}

	defer out.Close()

	// load the response body into a buffer
	buf := bytes.NewBuffer(nil)

	_, err = io.Copy(buf, res.Body)
	if err != nil {
		fmt.Println("Error copying response body", err)
		log.Fatal(err)
	}

	// write the buffer to the file
	_, err = out.Write(buf.Bytes())
	if err != nil {
		fmt.Println("Error writing to file", err)
		log.Fatal(err)
	}

	// read the response body
	agenciesResponse := &AgencyResponse{}

	// unmarshal the response body from the buffer
	err = json.NewDecoder(buf).Decode(agenciesResponse)
	if err != nil {
		fmt.Println("Error unmarshaling response body", err)
		log.Fatal(err)
	}

	os.MkdirAll("csv_output", 0755)

	outputFile := "csv_output/titles_all.csv"

	out, err = os.Create(outputFile)
	if err != nil {
		fmt.Println("Error creating file", err)
		log.Fatal(err)
	}

	defer out.Close()
	// write the header to the file
	_, err = out.WriteString("Agency,Display Name,Short Name,Year,Amount\n")
	if err != nil {
		fmt.Println("Error writing to file", err)
		log.Fatal(err)
	}

	outTotals, err := os.Create("csv_output/titles_totals.csv")
	if err != nil {
		fmt.Println("Error creating file", err)
		log.Fatal(err)
	}
	defer outTotals.Close()
	// write the header to the file
	_, err = out.WriteString("Year,Amount\n")
	if err != nil {
		fmt.Println("Error writing to file", err)
		log.Fatal(err)
	}

	for _, dir := range titlesDirs {
		fmt.Println("Processing directory", dir.path)
		titlesDir := dir.path

		// load all titles
		totalWC := 0
		titles := map[int]*ecfr.ECFR{}
		for i := 1; i <= 50; i++ {
			if i == 35 {
				fmt.Println("Skipping [title=" + strconv.Itoa(i) + "]")
				continue
			}

			filename := filepath.Join(titlesDir, "raw", fmt.Sprintf("title_%d.xml", i))
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				fmt.Println("File does not exist", filename)
				continue
			}

			titles[i], err = ecfr.NewECFR(filename)
			if err != nil {
				fmt.Println("Error loading title", i, err)
				continue
			}

			// count the number of words in the title
			c, err := titles[i].CountAllWords()
			if err != nil {
				fmt.Println("Error counting words in title", i, err)
				continue
			}
			totalWC += c
		}
		// add line to csv for all titles
		_, err = outTotals.WriteString(fmt.Sprintf("%s,%d\n", dir.year, totalWC))
		if err != nil {
			fmt.Println("Error writing to file", dir.year, err)
			log.Fatal(err)
		}

		// loop through the agencies
		for _, agency := range agenciesResponse.Agencies {
			// count the number of words in a agencies regulations files
			totalWords := 0

			for _, ref := range agency.CfrReferences {
				t, ok := titles[ref.Title]
				if !ok {
					fmt.Println("Skipping [title="+strconv.Itoa(ref.Title)+"]", "no title for", agency.Name, ref)
					continue
				}

				if ref.Subtitle != "" {
					c, err := t.CountWordsSubtitle(ref.Subtitle)
					if err != nil {
						fmt.Println("Error counting words in subtitle", ref.Subtitle, err)
						continue
					}

					totalWords += c
					continue
				}

				if ref.Part != "" {
					c, err := t.CountWordsPart(ref.Chapter, ref.Part)
					if err != nil {
						fmt.Println("Error counting words in part", ref.Part, err)
						continue
					}
					totalWords += c
					continue
				}

				if ref.Chapter != "" {
					c, err := t.CountWordsChapter(ref.Chapter)
					if err != nil {
						fmt.Println("Error counting words in chapter", ref.Chapter, err)
						continue
					}

					totalWords += c
					continue
				}

				// some kind of error happened
				fmt.Println("Skipping [title="+strconv.Itoa(ref.Title)+"]", "no subtitle, part, or chapter for", agency.Name, ref)
			}

			// for the child agencies
			for _, child := range agency.Children {
				// count the number of words in a agencies regulations files
				for _, ref := range child.CfrReferences {
					t, ok := titles[ref.Title]
					if !ok {
						fmt.Println("Skipping [title="+strconv.Itoa(ref.Title)+"]", "no title for", agency.Name, ref)
						continue
					}
					if ref.Subtitle != "" {
						c, err := t.CountWordsSubtitle(ref.Subtitle)
						if err != nil {
							fmt.Println("Error counting words in subtitle", ref.Subtitle, err)
							continue
						}
						totalWords += c
						continue
					}
					if ref.Part != "" {
						c, err := t.CountWordsPart(ref.Chapter, ref.Part)
						if err != nil {
							fmt.Println("Error counting words in part", ref.Part, err)
							continue
						}
						totalWords += c
						continue
					}
					if ref.Chapter != "" {
						c, err := t.CountWordsChapter(ref.Chapter)
						if err != nil {
							fmt.Println("Error counting words in chapter", ref.Chapter, err)
							continue
						}
						totalWords += c
						continue
					}
					// some kind of error happened
					fmt.Println("Skipping [title="+strconv.Itoa(ref.Title)+"]", "no subtitle, part, or chapter for", agency.Name, ref)
				}
			}

			// write the agency name and total words to the file in csv format
			_, err = out.WriteString(fmt.Sprintf("%s,%s,%s,%s,%d\n", agency.Name, agency.DisplayName, agency.ShortName, dir.year, totalWords))
			if err != nil {
				fmt.Println("Error writing to file", dir.year, err)
				log.Fatal(err)
			}
			// fmt.Println("Agency", agency.Name, "total words", totalWords)
		}

		log.Println("Done", dir.path)
	}

}
