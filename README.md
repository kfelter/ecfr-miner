# eCFR-Miner

This application will mine all of the eCFR regulations out of the public eCFR api, parse the XML and produce a useful dashboard that can be viewed on a web browser.

The purpose of this application is to provide insights into the US regulations in a data driven application. This should be used to justify regulatory changes, and to provide a better understanding of the regulations that govern our lives.

## Quickstart

1. Run the miner to download all the xml files. This is necessary because the eCFR api does not correctly support filtering by chapter for a title, so we have to download all 50 titles and parse them locally.

```bash
go run cmd/miner/main.go
```

2. Run the agency word count script to get the word counts for each agency.

```bash
go run cmd/agency-wc/main.go
```

## Code

### CMD/MINER
This golang script mines the data from the eCFR api and stores it in a local directory. It is a simple script that just downloads the files. It has 2 modes since it seems like some of the titles are not hydrated in the nginx cache.

The first mode attempts to download all the titles, it tries only once per file and gets a lot of errors with gateway timeouts. The second mode will follow the first and check for missing files that need to be downloaded.

Its best to run both at the same time so that you can get all the files quickly while making sure you get all the data.


### CMD/AGENCY-WC

This golang script gets the json data that relates each title, chapter, part, and subpart to the agency. Those references are parsed and used to lookup the words for that regulation. All the references are checked and added up to get the total word count for each agency. Some agencies have child agencies and those are included in the count. Some agencies have references to entire subparts that include many chapters and some references have # ecfr-miner


## CMD/REFINE

Dont use this.


This golang script was used during development to split up all the titles into their chapters. I learned a lot while making this but ultimately rewrote its functionality into the ecfr package and now it just uses the functions to count the words by loading the entire title instead of relying on the split files. This cleans things up a bit and gets rid of the derived data.
