# eCFR-Miner

This application will mine all of the eCFR regulations out of the public eCFR api, parse the XML and produce a useful dashboard that can be viewed on a web browser.

The purpose of this application is to provide insights into the US regulations in a data driven application. This should be used to justify regulatory changes, and to provide a better understanding of the regulations that govern our lives.

In this project I would have liked to include all data that goes back to 1996, but I was only able to scrape the data from the public api going back to 2017, some worked for 2016 but not enough for the whole set of titles. This data exists on other sites like https://www.govinfo.gov/app/collection/cfr/2024/ but it looks like they changed the XML format from the eCFR and it would have required reverse engineering another parser. Another challenge I faced was that the public api claimed to support fetching individual titles but it did not work correctly. I had to download all 50 titles and parse them locally. Some of the titles did not download from the api, facing errors like 504 Gateway Timeout. Notably title 40, I was only able to get data from 2025 on title 40 and copied that file accross all past years. This should be fixed by downloading the regulations from alternative sources and parsing it differently.

If I spent more time on this project I would download all data from 1996 and place it into a SQL database for improved query performance, I would probably implement LLM summaries, and use LLMs to decipher some of the legal speak, while bringing in context from any regulations that are referenced in the regulation. This would be a useful tool for reducing regulations without risking the safety of the public. To scale this application I would have backend servers to handle requests for json data that the static react frontend makes, along with caching the responses and intellegent indexes in the SQL database.

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
