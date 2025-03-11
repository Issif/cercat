package collector

import (
	"cercat/config"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

type Report struct {
	CountLinks    int
	CountAnchors  int
	CountLinks404 int
	Headers       []string
}

func Visit(domain string, u *url.URL) *Report {
	collector := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(2),
		colly.URLFilters(
			regexp.MustCompile(".+"+strings.Split(u.Hostname(), ".")[0]+".+$"),
		),
	)

	extensions.RandomUserAgent(collector)

	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 100,
		RandomDelay: 1 * time.Second,
	})

	report := new(Report)

	collector.OnError(func(response *colly.Response, err error) {
		if response.StatusCode == http.StatusNotFound {
			report.CountLinks404++
		}
	})

	collector.OnHTML("a[href]", func(element *colly.HTMLElement) {
		link := element.Attr("href")

		report.CountLinks++

		if link == "#" || len(link) == 0 {
			report.CountAnchors++
		}
		_ = element.Request.Visit(link)
	})

	collector.OnResponse(func(r *colly.Response) {
		report.Headers = append(report.Headers, r.Headers.Values("X-Powered-By")...)
	})

	collector.Visit(u.String())
	collector.Wait()

	return report
}

func (r *Report) GetStatus() string {
	cfg := config.GetConfig()
	if r.CountLinks > 0 {
		if 100*(r.CountAnchors+r.CountLinks404)/r.CountLinks > cfg.WrongLinksThreshold {
			return "confirmed"
		}
	}
	for _, i := range r.Headers {
		if strings.Contains(strings.ToLower(i), "plesk") {
			return "suspicious"
		}
	}
	return ""
}
