package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/chooban/progger/download/api"
	"github.com/go-logr/logr"
	"github.com/playwright-community/playwright-go"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var signinUrl = "https://shop.2000ad.com/account/sign-in"
var accountUrl = "https://shop.2000ad.com/account"
var listUrl = "https://shop.2000ad.com/account/downloads?sort-by=released&direction=desc"
var downloadPageUrl = "https://shop.2000ad.com/account/downloads?sort-by=granted&direction=desc&page=%d"

func Login(ctx context.Context, bContext playwright.BrowserContext, username, password string) (err error) {
	logger := logr.FromContextOrDiscard(ctx)
	assertions := playwright.NewPlaywrightAssertions()

	page, err := bContext.NewPage()
	if err != nil {
		return
	}
	if _, err = page.Goto(signinUrl); err != nil {
		return
	}

	if page.URL() == accountUrl {
		logger.Info("Skipping login procedure")
		return
	}
	var emailInput, passwordInput, loginButton playwright.Locator
	emailInput = page.GetByLabel("Email Address")
	passwordInput = page.GetByLabel("Password")
	loginButton = page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Sign In"})

	for _, v := range []playwright.Locator{emailInput, passwordInput, loginButton} {
		if err = assertions.Locator(v).ToBeVisible(); err != nil {
			return
		}
	}

	if err = emailInput.Fill(username); err != nil {
		return
	}
	if err = passwordInput.Fill(password); err != nil {
		return
	}

	if err = loginButton.Click(); err != nil {
		return
	}

	if _, err = page.ExpectEvent("navigated", func() error {
		currentUrl := page.URL()
		logger.Info("Current URL: " + currentUrl)
		if currentUrl != accountUrl {
			logger.Info("Looks like login failed")
			return errors.New("login failed")
		}
		return nil
	}); err != nil {
		logger.Info("Returning an error")
		return err
	}
	logger.Info("Login succeeded")
	return
}

func getPage(ctx context.Context, bContext playwright.BrowserContext) (playwright.Page, error) {
	logger := logr.FromContextOrDiscard(ctx)
	page, err := bContext.NewPage()
	if err != nil {
		return nil, err
	}
	r, _ := regexp.Compile("png|jpg|gif|woff|css")
	err = page.Route(r, func(route playwright.Route) {
		route.Abort()
	})
	if err != nil {
		logger.Error(err, "Could not set Route intercept")
	}
	return page, nil
}

func pageDownloader(ctx context.Context, bContext playwright.BrowserContext, pageNumber int) []api.DigitalComic {
	logger := logr.FromContextOrDiscard(ctx)
	url := listUrl + fmt.Sprintf("&page=%d", pageNumber)
	page, _ := getPage(ctx, bContext)
	start := time.Now()
	progs := make([]api.DigitalComic, 0)
	if _, err := page.Goto(url); err != nil {
		logger.Error(err, "Failed to load page", "url", downloadPageUrl)
	} else {
		logger.Info("Downloaded page", "duration", time.Since(start))
		if newProgs, err := extractProgsFromPage(logger, page); err == nil {
			logger.Info("Found new progs", "count", len(newProgs))
			progs = newProgs
		}
	}
	return progs
}

func ListProgs(ctx context.Context, bContext playwright.BrowserContext) (progs []api.DigitalComic, err error) {
	logger := logr.FromContextOrDiscard(ctx)
	page, _ := bContext.NewPage()
	if _, err := page.Goto(listUrl); err != nil {
		logger.Error(err, "Failed to load page", "url", listUrl)
		return progs, err
	}

	links := page.Locator("ul.pagination").GetByRole("link").Filter(playwright.LocatorFilterOptions{HasNotText: "Next"}).Last()
	maxPageText, err := links.InnerText()
	if err != nil {
		logger.Error(err, "could not get text of last link")
		return
	}
	logger.Info("Converting max page text", "text", maxPageText)
	maxPage, _ := strconv.Atoi(maxPageText)

	var wg sync.WaitGroup
	var mu sync.Mutex

	logger.Info("Found max page count", "max_page", maxPage)

	progs = make([]api.DigitalComic, 0, maxPage*10)
	for p := range maxPage {
		if p == 0 {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			_progs := pageDownloader(ctx, bContext, p)
			mu.Lock()
			defer mu.Unlock()
			progs = append(progs, _progs...)
		}()
	}

	wg.Wait()

	return
}

func Download(ctx context.Context, bContext playwright.BrowserContext, comic api.DigitalComic) (string, error) {
	logger := logr.FromContextOrDiscard(ctx)
	page, err := bContext.NewPage()
	if err != nil {
		return "", fmt.Errorf("could not open page %g", err)
	}
	download, err := page.ExpectDownload(func() error {
		// Weirdly, we ignore the errors because Playwright now considers a navigation
		// that turns into a download to sometimes be an error
		page.Goto(comic.Downloads[api.Pdf])
		return nil
	}, playwright.PageExpectDownloadOptions{})
	if err != nil {
		logger.Error(err, "Failed to download")
		return "", fmt.Errorf("failed to get a download %g", err)
	}

	path, err := download.Path()
	if err != nil {
		logger.Error(err, "Failed to download")
		return "", fmt.Errorf("no path to file returned %g", err)
	}
	logger.Info(fmt.Sprintf("Path is %s", path))

	return path, nil
}

func extractProgsFromPage(logger logr.Logger, page playwright.Page) ([]api.DigitalComic, error) {
	titleMatcher, err := regexp.Compile(`(?si)2000\s+AD\s+prog\s+\d{1,}`)
	issueNumberMatcher, _ := regexp.Compile(`(?si)PRG(?P<Issue>\d{1,})D`)
	titleFilter := playwright.LocatorFilterOptions{HasText: titleMatcher}
	if err != nil {
		return []api.DigitalComic{}, err
	}
	locators, err := page.GetByRole("listitem").Filter(titleFilter).All()
	if err != nil {
		return []api.DigitalComic{}, err
	}
	logger.Info("Found listitems", "count", len(locators), "page", page.URL())
	progs := make([]api.DigitalComic, len(locators))
	for i, v := range locators {
		productUrl, _ := v.GetByRole("link").Filter(titleFilter).First().GetAttribute("href")

		m := issueNumberMatcher.FindStringSubmatch(productUrl)
		issueNumberRaw := m[issueNumberMatcher.SubexpIndex("Issue")]
		issueNumber, _ := strconv.Atoi(issueNumberRaw)

		pdfForm := v.Locator("form").Filter(playwright.LocatorFilterOptions{HasText: "PDF"})
		pdfUrl, _ := pdfForm.GetAttribute("action")

		cbzForm := v.Locator("form").Filter(playwright.LocatorFilterOptions{HasText: "CBZ"})
		cbzUrl, _ := cbzForm.GetAttribute("action")

		progs[i] = api.DigitalComic{
			Url:         productUrl,
			IssueNumber: issueNumber,
			Downloads: map[api.FileType]string{
				api.Pdf: pdfUrl,
				api.Cbz: cbzUrl,
			},
		}
	}
	return progs, nil
}
