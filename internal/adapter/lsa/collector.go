package lsa

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/ecfr"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/vertexai"
	"github.com/xai/ecfr-dereg-dashboard/internal/domain"
	"golang.org/x/net/html"
)

type Collector struct {
	ecfr    *ecfr.Client
	vertex  *vertexai.Client
	baseURL string
	client  *http.Client
}

func NewCollector(ecfr *ecfr.Client, vertex *vertexai.Client) *Collector {
	return &Collector{
		ecfr:    ecfr,
		vertex:  vertex,
		baseURL: "https://www.govinfo.gov/app/collection/LSA",
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Collector) CollectForTitle(ctx context.Context, title string, since time.Time) (domain.LSAActivity, error) {
	url := c.baseURL + "/" + time.Now().Format("2006/01")
	resp, err := c.client.Get(url)
	if err != nil {
		return domain.LSAActivity{}, err
	}
	defer resp.Body.Close()

	// Simple HTML parsing to count keywords if structure is unknown or variable
	// In a real scenario, we'd use goquery or similar.
	// Here we just count occurrences in the text content.
	proposals := 0
	amendments := 0
	finals := 0

	tokenizer := html.NewTokenizer(resp.Body)
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}
		if tt == html.TextToken {
			text := string(tokenizer.Text())
			if strings.Contains(strings.ToLower(text), "proposal") {
				proposals++
			}
			if strings.Contains(strings.ToLower(text), "amendment") {
				amendments++
			}
			if strings.Contains(strings.ToLower(text), "final rule") {
				finals++
			}
		}
	}

	return domain.LSAActivity{
		KeyKind:         "title",
		Key:             title,
		SinceRevDate:    since,
		ProposalsCount:  proposals,
		AmendmentsCount: amendments,
		FinalsCount:     finals,
		CapturedAt:      time.Now(),
		SourceHint:      "govinfo-parse",
	}, nil
}

func (c *Collector) fallbackWithVertex(ctx context.Context, title string, since time.Time) (domain.LSAActivity, error) {
	prompt := "Search GovInfo for LSA changes in title " + title + " since " + since.Format("2006-01-02") + ". Extract counts: proposals, amendments, finals."
	tools := []vertexai.Tool{}
	_, err := c.vertex.CallWithTools(ctx, prompt, tools)
	if err != nil {
		return domain.LSAActivity{}, err
	}

	return domain.LSAActivity{
		KeyKind:         "title",
		Key:             title,
		SinceRevDate:    since,
		ProposalsCount:  0,
		AmendmentsCount: 0,
		FinalsCount:     0,
		CapturedAt:      time.Now(),
		SourceHint:      "vertex-tool",
	}, nil
}
