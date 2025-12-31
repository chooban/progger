package download

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/playwright-community/playwright-go"
)

var signinUrl = "https://shop.2000ad.com/account/sign-in"
var accountUrl = "https://shop.2000ad.com/account"
var listUrl = "https://shop.2000ad.com/account/downloads?sort-by=released&direction=desc"

//var downloadPageUrl = "https://shop.2000ad.com/account/downloads?sort-by=granted&direction=desc&page=%d"

var (
	resourceBlockRegex = regexp.MustCompile(`png|jpg|gif|woff|css`)
	issueNumberRegex   = regexp.MustCompile(`(?si)(PRG|MEG)(?P<Issue>\d{1,})D`)
	titleRegex         = regexp.MustCompile(`(?si)2000\s+AD\s+prog\s+\d{1,}|Judge\s+Dredd\s+Megazine\s+\d{1,}`)
	ordinalDateRegex   = regexp.MustCompile(`(\d+)(st|rd|th|nd)`)
)

func Login(ctx context.Context, bContext playwright.BrowserContext, username, password string) (err error) {
	logger := logr.FromContextOrDiscard(ctx)
	assertions := playwright.NewPlaywrightAssertions()

	if username == "" {
		return errors.New("login failed due to empty username")
	}
	if password == "" {
		return errors.New("login failed due to empty password")
	}

	page, err := getPage(ctx, bContext)
	defer func() {
		if err := page.Close(); err != nil {
			logger.Error(err, "Failed to close page")
			return
		}
	}()

	if err != nil {
		return
	}
	if _, err = page.Goto(accountUrl); err != nil {
		logger.Info("Failed to navigate to login page")
		return
	}

	var doLogin = func() (err error) {
		if page.URL() == accountUrl {
			logger.Info("Skipping login procedure")
			return
		}

		logger.Info(fmt.Sprintf("Trying to log in %s", page.URL()))

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

		var timeout float64 = 5000
		if err := page.WaitForURL(accountUrl, playwright.PageWaitForURLOptions{Timeout: &timeout}); err != nil {
			logger.Info("Looks like login failed", "current_url", page.URL(), "expected_url", accountUrl)
			logger.V(1).Info(page.Content())
			return errors.New("login failed")
		}

		return
	}

	for attempts := 0; attempts < 3; attempts++ {
		err = doLogin()
		if err == nil {
			logger.Info("Login succeeded")
			break
		}
		logger.Error(err, "Failed to login")
		time.Sleep(3 * time.Second)
		if _, err := page.Goto(signinUrl); err != nil {
			logger.Error(err, "Failed to open page", "url", signinUrl)
		}
	}

	return
}

func getPage(ctx context.Context, bContext playwright.BrowserContext) (page playwright.Page, err error) {
	logger := logr.FromContextOrDiscard(ctx)
	if page, err = bContext.NewPage(); err != nil {
		logger.Error(err, "Failed to create page")
		return
	}
	err = page.Route(resourceBlockRegex, func(route playwright.Route) {
		if err := route.Abort(); err != nil {
			logger.Error(err, "Failed to abort route")
		}
	})
	if err != nil {
		logger.Error(err, "Could not set Route intercept")
	}
	return
}

func pageDownloader(ctx context.Context, bContext playwright.BrowserContext, pageNumber int) []DigitalComic {
	logger := logr.FromContextOrDiscard(ctx)
	url := listUrl + fmt.Sprintf("&page=%d", pageNumber)
	logger.Info(fmt.Sprintf("Downloading page %s", url))
	var page playwright.Page
	var err error
	progs := make([]DigitalComic, 0)
	if page, err = getPage(ctx, bContext); err != nil {
		logger.Error(err, "Failed to get page from list")
		return progs
	}
	defer func() {
		if page.IsClosed() {
			return
		}
		if err := page.Close(); err != nil {
			logger.Error(err, "Failed to close page")
			return
		}
	}()
	start := time.Now()
	if _, err := page.Goto(url); err != nil {
		logger.Error(err, "Failed to load page", "url", url)
	} else {
		logger.Info("Downloaded page", "duration", time.Since(start))
		if newProgs, err := extractProgsFromPage(logger, page); err == nil {
			logger.Info("Found new progs", "count", len(newProgs))
			progs = newProgs
		} else {
			logger.Error(err, "Failed to extract progs")
		}
	}
	return progs
}

func listIssuesOnPage(ctx context.Context, bContext playwright.BrowserContext, pageNumber int) (issues []DigitalComic, err error) {
	issues = pageDownloader(ctx, bContext, pageNumber)

	return
}

func listProgs(ctx context.Context, bContext playwright.BrowserContext, latestOnly bool) (allProgs []DigitalComic, err error) {
	logger := logr.FromContextOrDiscard(ctx)

	maxPage := 1
	if !latestOnly {
		logger.Info(fmt.Sprintf("Listing all progs..."))
		page, err := getPage(ctx, bContext)
		if err != nil {
			logger.Error(err, "Failed to get page")
			return nil, err
		}
		if _, err := page.Goto(listUrl); err != nil {
			logger.Error(err, "Failed to load list page", "url", listUrl)
			return allProgs, err
		}

		links := page.Locator("ul.pagination").GetByRole("link").Filter(playwright.LocatorFilterOptions{HasNotText: "Next"}).Last()
		if maxPageText, _err := links.InnerText(); _err != nil {
			logger.Error(err, "could not get text of last link")
			return nil, _err
		} else {
			maxPage, err = strconv.Atoi(maxPageText)
			if err != nil {
				logger.Error(err, "Failed to parse max page number")
				return nil, fmt.Errorf("failed to parse max page: %w", err)
			}
		}
		page.Close()
	} else {
		logger.Info("Only retrieving most recent issues")
	}

	const maxWorkers = 5
	semaphore := make(chan struct{}, maxWorkers)

	var wg sync.WaitGroup
	var mu sync.Mutex

	allProgs = make([]DigitalComic, 0, maxPage*10)
	for p := range maxPage {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire

		go func(pageNum int) {
			defer func() {
				<-semaphore // Release
				wg.Done()
			}()

			progsFromPage := pageDownloader(ctx, bContext, pageNum)
			mu.Lock()
			allProgs = append(allProgs, progsFromPage...)
			mu.Unlock()
		}(p + 1) // Fix loop variable capture
	}

	wg.Wait()

	return
}

func downloadComic(ctx context.Context, bContext playwright.BrowserContext, comic DigitalComic) (string, error) {
	logger := logr.FromContextOrDiscard(ctx)
	page, err := getPage(ctx, bContext)
	if err != nil {
		return "", fmt.Errorf("could not open page: %w", err)
	}
	expectDownload, err := page.ExpectDownload(func() error {
		// Weirdly, we ignore the errors because Playwright now considers a navigation
		// that turns into a downloadComic to sometimes be an error
		page.Goto(comic.Downloads[Pdf])
		return nil
	}, playwright.PageExpectDownloadOptions{})
	if err != nil {
		logger.Error(err, "Failed to download an issue")
		return "", fmt.Errorf("failed to download an issue: %w", err)
	}

	path, err := expectDownload.Path()
	if err != nil {
		logger.Error(err, "No path found for downloaded issue")
		return "", fmt.Errorf("no path to file returned: %w", err)
	}
	logger.Info(fmt.Sprintf("Path is %s", path))

	return path, nil
}

func extractProgsFromPage(logger logr.Logger, page playwright.Page) ([]DigitalComic, error) {
	titleFilter := playwright.LocatorFilterOptions{HasText: titleRegex}

	locators, err := page.GetByRole("listitem").Filter(titleFilter).All()
	if err != nil {
		return []DigitalComic{}, err
	}
	logger.Info("Found listitems", "count", len(locators), "page", page.URL())
	progs := make([]DigitalComic, 0, len(locators))
	for _, v := range locators {
		productUrl, err := v.GetByRole("link").Filter(titleFilter).First().GetAttribute("href")
		if err != nil {
			logger.V(1).Info("Failed to get product URL, skipping item", "error", err)
			continue
		}

		m := issueNumberRegex.FindStringSubmatch(productUrl)
		if m == nil {
			logger.V(1).Info("Failed to match issue number in URL, skipping item", "url", productUrl)
			continue
		}
		issueNumberRaw := m[issueNumberRegex.SubexpIndex("Issue")]
		issueNumber, err := strconv.Atoi(issueNumberRaw)
		if err != nil {
			logger.V(1).Info("Failed to parse issue number, skipping item", "raw", issueNumberRaw, "error", err)
			continue
		}

		pdfForm := v.Locator("form").Filter(playwright.LocatorFilterOptions{HasText: "PDF"})
		pdfUrl, err := pdfForm.GetAttribute("action")
		if err != nil {
			logger.V(1).Info("Failed to get PDF URL", "error", err)
			pdfUrl = ""
		}

		cbzForm := v.Locator("form").Filter(playwright.LocatorFilterOptions{HasText: "CBZ"})
		cbzUrl, err := cbzForm.GetAttribute("action")
		if err != nil {
			logger.V(1).Info("Failed to get CBZ URL", "error", err)
			cbzUrl = ""
		}

		issueDate := v.Locator("[class=subheader]").First()
		dateString, err := issueDate.InnerText()

		dateString = ordinalDateRegex.ReplaceAllString(dateString, "$1")
		d, err := time.Parse("2 January 2006", dateString)
		if err != nil {
			logger.Error(err, "could not get date")
		}
		title := v.Locator("h2").Filter(titleFilter).First()
		titleText, err := title.InnerText()
		if err != nil {
			logger.V(1).Info("Failed to get title text, skipping item", "error", err)
			continue
		}
		publication := "2000AD"
		if strings.Contains(titleText, "Megazine") {
			publication = "Megazine"
		}
		progs = append(progs, DigitalComic{
			Publication: publication,
			Url:         productUrl,
			IssueNumber: issueNumber,
			IssueDate:   d.Format("2006-01-02"),
			Downloads: map[FileType]string{
				Pdf: pdfUrl,
				Cbz: cbzUrl,
			},
		})
	}
	return progs, nil
}
