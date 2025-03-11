package utils

import (
	"net/http"
	"net/url"
)

// GetFinaleURL checks if the website is online and give the URL after possible redirections
func GetFinaleURL(domain string) *url.URL {
	res, err := http.Get("https://" + domain)
	if err != nil {
		return nil
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil
	}

	return res.Request.URL
}
