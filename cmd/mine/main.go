package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	outputDir     = "titles_" + fmt.Sprintf("%d/raw", time.Now().Unix())
	concurrency   = flag.Int("c", 1, "number of concurrent requests") // by default download all chapters
	dateCheckMode = flag.Bool("d", false, "check if the date is valid")
)

type Miner struct {
	maxConcurrency int
	sem            chan struct{}
	wg             sync.WaitGroup
}

func NewMiner(maxConcurrency int) *Miner {
	return &Miner{
		maxConcurrency: maxConcurrency,
		sem:            make(chan struct{}, maxConcurrency),
	}
}

func (m *Miner) AddJob(job func() error) {
	m.wg.Add(1)
	m.sem <- struct{}{}
	go func() {
		defer m.wg.Done()
		err := job()
		if err != nil {
			log.Println(err)
		}

		if *dateCheckMode {
			for i := 0; i < 1 && err != nil; i++ {
				err = job()
				if err == nil {
					break
				}
				log.Println(err)
			}
		}

		<-m.sem
	}()
}

func (m *Miner) Wait() {
	m.wg.Wait()
}

func main() {
	flag.Parse()

	miner := NewMiner(*concurrency)

	tstart := time.Now()

	for i := 0; i <= 8; i++ {
		year := 2025 - i

		outputDir = "titles_" + strconv.Itoa(year) + "/raw"

		date := fmt.Sprintf("%d-02-07", year)

		// create the output directory
		err := os.MkdirAll(outputDir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		// download all chapters 1-50 with concurrency
		// curl -X GET "https://www.ecfr.gov/api/versioner/v1/full/2025-02-07/title-22.xml" -H "accept: application/xml"
		for i := 1; i <= 50; i++ {
			f := func() error { return downloadChapter(date, i) }
			miner.AddJob(f)
		}

		miner.Wait()
	}

	fmt.Println("All downloads complete", time.Since(tstart))

}

func downloadChapter(date string, i int) error {
	// skip title 35
	if i == 35 {
		fmt.Println("Skipping [title=" + strconv.Itoa(i) + "]")
		return nil
	}

	fname := outputDir + "/title_" + strconv.Itoa(i) + ".xml"
	// check if the file exists first
	if _, err := os.Stat(fname); err == nil {
		fmt.Println("File already exists", fname)
		if *dateCheckMode {
			if fileInfo, err := os.Stat(fname); err == nil && fileInfo.Size() > 0 {
				return nil
			}
		} else {
			return nil
		}
	}

	// save the file to the output directory
	out, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Download staring [title=" + strconv.Itoa(i) + "]")

	res, err := http.Get("https://www.ecfr.gov/api/versioner/v1/full/" + date + "/title-" + strconv.Itoa(i) + ".xml")
	if err != nil {
		fmt.Println("Error downloading [title="+strconv.Itoa(i)+"]", err)
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		fmt.Println("Error downloading [title="+strconv.Itoa(i)+"]", res.Status)
		return fmt.Errorf("error downloading [title=%d] %s", i, res.Status)
	}

	defer out.Close()

	_, err = out.ReadFrom(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Download complete [title=" + strconv.Itoa(i) + "]")
	return nil
}
