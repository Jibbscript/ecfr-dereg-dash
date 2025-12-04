package lsa

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/domain"
)

// FederalRegisterResponse represents the API response from federalregister.gov
type FederalRegisterResponse struct {
	Count       int                    `json:"count"`
	Description string                 `json:"description"`
	TotalPages  int                    `json:"total_pages"`
	NextPageURL string                 `json:"next_page_url"`
	Results     []FederalRegisterDoc   `json:"results"`
	Facets      *FederalRegisterFacets `json:"facets,omitempty"`
}

type FederalRegisterDoc struct {
	DocumentNumber  string   `json:"document_number"`
	Type            string   `json:"type"`
	Title           string   `json:"title"`
	PublicationDate string   `json:"publication_date"`
	AgencyNames     []string `json:"agency_names"`
}

type FederalRegisterFacets struct {
	Agency map[string]int `json:"agency"`
}

// AgencyInfo represents agency data from the Federal Register agencies endpoint
type AgencyInfo struct {
	ID         int    `json:"id"`
	ParentID   *int   `json:"parent_id"`
	Name       string `json:"name"`
	ShortName  string `json:"short_name"`
	Slug       string `json:"slug"`
	URL        string `json:"url"`
	JSONurl    string `json:"json_url"`
	RecentDocs string `json:"recent_articles_url"`
}

type Collector struct {
	client *http.Client
	// Cache for agency LSA data
	agencyLSAData map[string]domain.AgencyLSA
	// Cache for agency info from Federal Register
	agencyInfo map[string]AgencyInfo
}

func NewCollector() *Collector {
	return &Collector{
		client:        &http.Client{Timeout: 60 * time.Second},
		agencyLSAData: make(map[string]domain.AgencyLSA),
		agencyInfo:    make(map[string]AgencyInfo),
	}
}

// CollectAgencyLSA fetches LSA data for all agencies from the Federal Register API.
// It queries recent regulatory documents (proposed rules, final rules, notices) per agency.
func (c *Collector) CollectAgencyLSA(ctx context.Context, agencies []string) ([]domain.AgencyLSA, error) {
	// Reset cache
	c.agencyLSAData = make(map[string]domain.AgencyLSA)

	snapshotDate := time.Now().Format("2006-01-02")
	capturedAt := time.Now()

	// Calculate date range: last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	var results []domain.AgencyLSA

	for _, agencySlug := range agencies {
		// Fetch counts for each document type
		proposedRules, err := c.fetchDocumentCount(ctx, agencySlug, "PRORULE", startDate, endDate)
		if err != nil {
			// Log but continue with other agencies
			proposedRules = 0
		}

		finalRules, err := c.fetchDocumentCount(ctx, agencySlug, "RULE", startDate, endDate)
		if err != nil {
			finalRules = 0
		}

		notices, err := c.fetchDocumentCount(ctx, agencySlug, "NOTICE", startDate, endDate)
		if err != nil {
			notices = 0
		}

		agencyLSA := domain.AgencyLSA{
			AgencyID:       agencySlug,
			AgencyName:     agencySlug, // Will be enriched later if we have agency info
			ProposedRules:  proposedRules,
			FinalRules:     finalRules,
			Notices:        notices,
			TotalDocuments: proposedRules + finalRules + notices,
			SnapshotDate:   snapshotDate,
			CapturedAt:     capturedAt,
			SourceHint:     "federalregister-api",
		}

		// Enrich with agency name if available
		if info, ok := c.agencyInfo[agencySlug]; ok {
			agencyLSA.AgencyName = info.Name
		}

		c.agencyLSAData[agencySlug] = agencyLSA
		results = append(results, agencyLSA)
	}

	return results, nil
}

// CollectAgencyLSABatch fetches LSA data for multiple agencies efficiently using faceted search.
// This is more efficient than individual queries when collecting for many agencies.
func (c *Collector) CollectAgencyLSABatch(ctx context.Context) ([]domain.AgencyLSA, error) {
	snapshotDate := time.Now().Format("2006-01-02")
	capturedAt := time.Now()

	// Calculate date range: last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	// Fetch document counts by type with agency facets
	proposedByAgency, err := c.fetchDocumentCountsWithFacets(ctx, "PRORULE", startDate, endDate)
	if err != nil {
		proposedByAgency = make(map[string]int)
	}

	finalByAgency, err := c.fetchDocumentCountsWithFacets(ctx, "RULE", startDate, endDate)
	if err != nil {
		finalByAgency = make(map[string]int)
	}

	noticesByAgency, err := c.fetchDocumentCountsWithFacets(ctx, "NOTICE", startDate, endDate)
	if err != nil {
		noticesByAgency = make(map[string]int)
	}

	// Merge all agency slugs
	allAgencies := make(map[string]bool)
	for slug := range proposedByAgency {
		allAgencies[slug] = true
	}
	for slug := range finalByAgency {
		allAgencies[slug] = true
	}
	for slug := range noticesByAgency {
		allAgencies[slug] = true
	}

	var results []domain.AgencyLSA
	for agencySlug := range allAgencies {
		proposed := proposedByAgency[agencySlug]
		final := finalByAgency[agencySlug]
		notices := noticesByAgency[agencySlug]

		agencyLSA := domain.AgencyLSA{
			AgencyID:       agencySlug,
			AgencyName:     agencySlug,
			ProposedRules:  proposed,
			FinalRules:     final,
			Notices:        notices,
			TotalDocuments: proposed + final + notices,
			SnapshotDate:   snapshotDate,
			CapturedAt:     capturedAt,
			SourceHint:     "federalregister-api-batch",
		}

		c.agencyLSAData[agencySlug] = agencyLSA
		results = append(results, agencyLSA)
	}

	return results, nil
}

// fetchDocumentCount queries the Federal Register API for document count by agency and type
func (c *Collector) fetchDocumentCount(ctx context.Context, agencySlug, docType string, startDate, endDate time.Time) (int, error) {
	// Build API URL
	// https://www.federalregister.gov/api/v1/documents.json?conditions[agencies][]=agency-slug&conditions[type][]=RULE&conditions[publication_date][gte]=2024-01-01
	baseURL := "https://www.federalregister.gov/api/v1/documents.json"

	params := url.Values{}
	params.Set("conditions[agencies][]", agencySlug)
	params.Set("conditions[type][]", docType)
	params.Set("conditions[publication_date][gte]", startDate.Format("2006-01-02"))
	params.Set("conditions[publication_date][lte]", endDate.Format("2006-01-02"))
	params.Set("per_page", "1") // We only need the count, not the documents
	params.Set("fields[]", "document_number")

	reqURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("federal register API returned %s", resp.Status)
	}

	var frResp FederalRegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&frResp); err != nil {
		return 0, err
	}

	return frResp.Count, nil
}

// fetchDocumentCountsWithFacets fetches document counts for all agencies using faceted search
func (c *Collector) fetchDocumentCountsWithFacets(ctx context.Context, docType string, startDate, endDate time.Time) (map[string]int, error) {
	baseURL := "https://www.federalregister.gov/api/v1/documents/facets/agency"

	params := url.Values{}
	params.Set("conditions[type][]", docType)
	params.Set("conditions[publication_date][gte]", startDate.Format("2006-01-02"))
	params.Set("conditions[publication_date][lte]", endDate.Format("2006-01-02"))

	reqURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("federal register facets API returned %s", resp.Status)
	}

	// The facets endpoint returns a map of agency slug -> count
	var facets map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&facets); err != nil {
		return nil, err
	}

	return facets, nil
}

// FetchFederalRegisterAgencies fetches the list of agencies from the Federal Register API
func (c *Collector) FetchFederalRegisterAgencies(ctx context.Context) ([]AgencyInfo, error) {
	reqURL := "https://www.federalregister.gov/api/v1/agencies.json"

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("federal register agencies API returned %s", resp.Status)
	}

	var agencies []AgencyInfo
	if err := json.NewDecoder(resp.Body).Decode(&agencies); err != nil {
		return nil, err
	}

	// Cache agency info
	for _, agency := range agencies {
		c.agencyInfo[agency.Slug] = agency
	}

	return agencies, nil
}

// GetAgencyLSA retrieves cached LSA data for a specific agency
func (c *Collector) GetAgencyLSA(agencySlug string) (domain.AgencyLSA, bool) {
	lsa, ok := c.agencyLSAData[agencySlug]
	return lsa, ok
}
