package internal

import (
	"context"
	"fmt"
	"github.com/chooban/progger/download/api"
	"github.com/go-logr/logr"
	"github.com/playwright-community/playwright-go"
	"regexp"
	"strconv"
)

var signinUrl = "https://shop.2000ad.com/account/sign-in"
var listUrl = "https://shop.2000ad.com/account/downloads?sort-by=released&direction=desc"

func Login(ctx context.Context, bContext playwright.BrowserContext, username, password string) (err error) {
	assertions := playwright.NewPlaywrightAssertions()

	page, err := bContext.NewPage()
	if err != nil {
		return
	}
	if _, err = page.Goto(signinUrl); err != nil {
		return
	}

	if page.URL() != signinUrl {
		// Presumably we're logged in?
		logger := logr.FromContextOrDiscard(ctx)
		logger.V(1).Info("Skipping login procedure")
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

	return
}

func ListProgs(ctx context.Context, bContext playwright.BrowserContext) (progs []api.DigitalComic, err error) {
	//assertions := playwright.NewPlaywrightAssertions()
	logger := logr.FromContextOrDiscard(ctx)
	page, err := bContext.NewPage()
	if err != nil {
		return
	}
	if resp, err := page.Goto(listUrl); err != nil {
		return progs, err
	} else {
		logger.V(1).Info("Response code", "response_code", resp.Status())
	}

	progs, _ = extractProgsFromPage(logger, page)

	// TODO: Iterate through for a full list
	//nextLink := page.GetByRole("link").Filter(playwright.LocatorFilterOptions{HasText: "Next"}).First()

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
	logger.Info("Found listitems", "count", len(locators))
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
