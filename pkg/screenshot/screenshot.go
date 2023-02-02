package screenshot

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
)

// TakeScreenshot takes a screenshot, uploads it and return the URL
func TakeScreenshot(domain string) string {
	url := getFinaleURL(domain)

	if url == "" {
		return ""
	}

	quality := 90

	opts := []chromedp.ExecAllocatorOption{}
	opts = append(opts, chromedp.DefaultExecAllocatorOptions[:]...)
	opts = append(opts,
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,
		chromedp.NoSandbox,
		chromedp.DisableGPU,
		chromedp.WindowSize(1920, 1080),
		chromedp.IgnoreCertErrors,
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	folder := "/tmp/"
	browserCtx, browserCancel := chromedp.NewContext(
		allocCtx,
		// chromedp.WithDebugf(log.Printf),
	)
	defer browserCancel()

	tabCtx, cancelTabCtx := context.WithTimeout(browserCtx, time.Duration(15)*time.Second)
	defer cancelTabCtx()

	var buf []byte
	err := chromedp.Run(
		tabCtx,
		chromedp.Tasks{
			chromedp.Navigate(url),
			chromedp.Sleep(time.Second * 3),
			chromedp.FullScreenshot(&buf, quality),
		},
	)
	if err != nil {
		log.Warnf("Can't take a screenshot of domain '%v': %v", domain, err)
		return ""
	}
	if err = ioutil.WriteFile(folder+domain+".png", buf, 0o644); err != nil {
		log.Warnf("Can't write the .png of the screenshot of domain '%v': %v", domain, err)
		return ""
	}
	log.Infof("Screenshot taken for domain '%v'", domain)

	file, err := os.Open(folder + domain + ".png")
	if err != nil {
		log.Warnf("Can't open the screenshot of domain '%v': %v\n", domain, err)
		return ""
	}
	// defer func() {
	// 	if err := os.Remove(folder + domain + ".png"); err != nil {
	// 		log.Warnf("Can't delete the screenshot file %v%v.png: %v\n", folder, domain, err)
	// 	}
	// }()
	defer file.Close()

	req, err := http.NewRequest("PUT", "https://transfer.sh/"+domain+".png", file)
	if err != nil {
		log.Warnf("Can't upload the screenshot of domain '%v': %v", domain, err)
		return ""
	}
	req.Header.Set("Content-Type", "image/png")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Warnf("Can't upload the screenshot of domain '%v': %v", domain, err)
		return ""
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return ""
	}

	message, _ := ioutil.ReadAll(res.Body)
	return string(message)
}

// check if the website is online and if a redirect is required
func getFinaleURL(domain string) string {
	res, err := http.Get("https://" + domain)
	if err != nil {
		return ""
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return ""
	}

	return res.Request.URL.String()
}
