package govinfo

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xai/ecfr-dereg-dashboard/internal/domain"
)

type Client struct {
	baseURL string
	client  *http.Client
	dataDir string
}

func NewClient(dataDir string) *Client {
	return &Client{
		baseURL: "https://www.govinfo.gov/bulkdata/json/ECFR",
		client:  &http.Client{Timeout: 10 * time.Minute},
		dataDir: dataDir,
	}
}

type IndexResponse struct {
	Files []IndexFile `json:"files"`
}

type IndexFile struct {
	Name     string `json:"name"`
	CFRTitle int    `json:"cfrTitle"`
	Link     string `json:"link"`
}

type TitleResponse struct {
	Files []TitleFile `json:"files"`
}

type TitleFile struct {
	Name          string `json:"name"`
	FileExtension string `json:"fileExtension"`
	Link          string `json:"link"`
}

// DownloadTitleXML downloads the latest XML for a given title using API discovery
func (c *Client) DownloadTitleXML(title int) (string, error) {
	// Step 1: Fetch main index
	indexURL := c.baseURL
	var indexResp IndexResponse
	if err := c.fetchJSON(indexURL, &indexResp); err != nil {
		return "", fmt.Errorf("failed to fetch index: %w", err)
	}

	// Step 2: Find link for title
	var titleLink string
	for _, f := range indexResp.Files {
		if f.CFRTitle == title {
			titleLink = f.Link
			break
		}
	}
	if titleLink == "" {
		return "", fmt.Errorf("title %d not found in index", title)
	}

	// Step 3: Fetch title details
	var titleResp TitleResponse
	if err := c.fetchJSON(titleLink, &titleResp); err != nil {
		return "", fmt.Errorf("failed to fetch title details: %w", err)
	}

	// Step 4: Find XML file link
	var xmlLink string
	var xmlName string
	for _, f := range titleResp.Files {
		if f.FileExtension == "xml" {
			xmlLink = f.Link
			xmlName = f.Name
			break
		}
	}
	if xmlLink == "" {
		return "", fmt.Errorf("XML file not found for title %d", title)
	}

	// Step 5: Download file
	path := filepath.Join(c.dataDir, "raw", xmlName)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}

	// Check if already exists
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	if err := c.downloadFile(xmlLink, path); err != nil {
		return "", err
	}

	return path, nil
}

func (c *Client) fetchJSON(url string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %s", resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func (c *Client) downloadFile(url, path string) error {
	resp, err := c.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: %s", resp.Status)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

// ParseTitleXML parses the XML file into sections
// Note: This function remains largely the same as before, assuming the XML structure is what we expect.
// We just need to ensure imports and package name are correct.
func (c *Client) ParseTitleXML(path string) ([]domain.Section, error) {
	// ... (rest of the parsing logic, which we can keep or re-implement if needed)
	// Since replace_file_content replaces the whole range, I need to include the parsing logic here.
	// I will copy the previous parsing logic.

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var sections []domain.Section
	decoder := xml.NewDecoder(f)

	// Simplified XML parsing logic.
	// Real eCFR XML is complex. We'll look for DIV8 (Section) nodes.
	// This is a simplified stream parser.

	var currentText strings.Builder
	var inSection bool
	var sectionID string
	var currentAgencyID string

	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch se := t.(type) {
		case xml.StartElement:
			if getAttr(se, "TYPE") == "CHAPTER" {
				currentAgencyID = getAttr(se, "N")
			}
			if se.Name.Local == "DIV8" {
				inSection = true
				sectionID = getAttr(se, "N")
				currentText.Reset()
			}
		case xml.CharData:
			if inSection {
				currentText.Write(se)
			}
		case xml.EndElement:
			if se.Name.Local == "DIV8" {
				inSection = false
				sections = append(sections, domain.Section{
					ID:       sectionID,
					Section:  sectionID,
					AgencyID: currentAgencyID,
					Text:     currentText.String(),
					// Other fields need more context or post-processing
				})
			}
		}
	}
	return sections, nil
}

func getAttr(se xml.StartElement, name string) string {
	for _, a := range se.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}
