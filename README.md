# Millipedes

## Description
My version of the crawler scrapes a website recursively starting from an entrypoint. It supports multiple workers so that it can scrape multiple links in concurrently.
Results are visualised using SigmaJS (http://sigmajs.org/). Once finished crawling the SigmaJS render starts a HTTP server which exposes the SigmaJS JSON representation via the `/data` endpoint. The browser is automatically opened pointing to the `/` endpoint, which shows the HTML page with the graph.
A ratelimiter is set to 200ms, so that HTTP requests can be limited.

## Running it
The simplest option to run it is to run:
```
make build-run
```

Note that by default I've set the website (entrypoint) to be `https://example.com`.

## Usage
```
Usage of ./crawler:
  -loglevel string
        log level (debug/info/warn/fatal (default "info")
  -queue int
        the queue size to store pending urls that need parsing (default 1000)
  -rate int
        the rate limiter interval in ms (default 200)
  -website string
        the website to be crawled (default "https://example.com/")
  -workers int
        the number of concurrent workers (default 100)
```

## Assumptions
- this is a tool to get the sitemap for a website and not a service running continuously
- when the website `https://example.com` is crawled, all its subdomains are too
- the program renders the results when finished crawling but also after cancelling the execution
- the `static` folder is at the same level of the binary

## Things I decided to skip
- since this is a tool, I haven't exposed any metrics
- I haven't used a Builder pattern to allow the user to switch between different implementations of the render
- testing the resulting structure from the conversion to a sigma object
- robots.txt files are ignored
- HTTP port configuration for SigmaJS is configurable in the constructor but not available for simplicity
- HTTP timeouts are statically configured
- didn't create a package so `static` folder needs to be at the same level as the binary
- it's not possible to move nodes of the SigmaJS graph