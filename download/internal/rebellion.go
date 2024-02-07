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
	//logger := logr.FromContextOrDiscard(ctx).V(1)

	page, err := bContext.NewPage()
	if err != nil {
		return
	}
	if _, err = page.Goto(signinUrl); err != nil {
		return
	}

	if page.URL() != signinUrl {
		// Presumably we're logged in?
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
	//logger.Info("Username is", "username", username)
	//logger.Info("Password is", "password", password)

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

	if title, err := page.Title(); err != nil {
		logger.Error(err, "Could not get title")
	} else {
		logger.Info(fmt.Sprintf("Current page title: %s", title))
	}

	progs, _ = extractProgsFromPage(logger, page)

	return
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
		logger.Info("Found a url", "url", productUrl)

		m := issueNumberMatcher.FindStringSubmatch(productUrl)
		issueNumberRaw := m[issueNumberMatcher.SubexpIndex("Issue")]
		issueNumber, _ := strconv.Atoi(issueNumberRaw)

		progs[i] = api.DigitalComic{Url: productUrl, IssueNumber: issueNumber}
	}
	return progs, nil
}
