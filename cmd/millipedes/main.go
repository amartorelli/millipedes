package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amartorelli/millipedes/pkg/crawler"
	"github.com/amartorelli/millipedes/pkg/crawler/fetcher"
	"github.com/amartorelli/millipedes/pkg/crawler/render"
	"github.com/amartorelli/millipedes/pkg/crawler/sitemap"

	"github.com/sirupsen/logrus"
)

func main() {
	// signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// flags
	website := flag.String("website", "https://example.com/", "the website to be crawled")
	workers := flag.Int("workers", 100, "the number of concurrent workers")
	queueLen := flag.Int("queue", 1000, "the queue size to store pending urls that need parsing")
	rate := flag.Int("rate", 200, "the rate limiter interval in ms")
	loglevel := flag.String("loglevel", "info", "log level (debug/info/warn/fatal")
	flag.Parse()

	// logging
	switch *loglevel {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	formatter := &logrus.TextFormatter{
		FullTimestamp: true,
	}
	logrus.SetFormatter(formatter)

	fetcher := fetcher.NewHTTPFetcher()
	sitemap := sitemap.NewMemorySitemap()
	c, err := crawler.NewCrawler(*website, *workers, *queueLen, *rate, fetcher, sitemap)
	if err != nil {
		logrus.Fatal(err)
	}

	c.Start()

	run := true
	for run {
		select {
		case sig := <-sigs:
			logrus.Infof("received signal %s, gracefully shutting down: processing all remaining elements in the queue before exiting\n", sig)
			run = false
		case <-time.Tick(1 * time.Second):
			if c.IsDone() {
				run = false
			}
		}
	}

	// show results
	c.Shutdown()
	logrus.Info("done")
	sm := c.Sitemap()
	r := render.NewSigmajsRender(":9876")
	r.UpdateSitemap(sm)
	err = r.Render()
	if err != nil {
		logrus.Fatal(err)
	}
}
