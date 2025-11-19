package govinfo

import (
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
		baseURL: "https://www.govinfo.gov/bulkdata/ECFR",
		client:  &http.Client{Timeout: 10 * time.Minute}, // Long timeout for large files
		dataDir: dataDir,
	}
}

// DownloadTitleXML downloads the latest XML for a given title
func (c *Client) DownloadTitleXML(title int) (string, error) {
	// Format: ECFR-{year}-Title-{title}.xml
	// We need to find the correct URL. For simplicity, we'll assume a structure or use a directory listing if possible.
	// GovInfo bulk data structure: /ECFR/{year}/Title-{title}/ECFR-{year}-Title-{title}.xml
	year := time.Now().Year()
	filename := fmt.Sprintf("ECFR-%d-Title-%d.xml", year, title)
	url := fmt.Sprintf("%s/%d/Title-%d/%s", c.baseURL, year, title, filename)

	path := filepath.Join(c.dataDir, "raw", filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}

	// Check if already exists
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	resp, err := c.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to download: %s", resp.Status)
	}

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return path, err
}

// ParseTitleXML parses the XML file into sections
func (c *Client) ParseTitleXML(path string) ([]domain.Section, error) {
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

	// var currentPart string
	// var currentSectionTitle string
	var currentText strings.Builder
	var inSection bool
	var sectionID string

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
			if se.Name.Local == "DIV8" {
				inSection = true
				sectionID = getAttr(se, "N")
				currentText.Reset()
			}
			if inSection && se.Name.Local == "HEAD" {
				// Capture title
			}
		case xml.CharData:
			if inSection {
				currentText.Write(se)
			}
		case xml.EndElement:
			if se.Name.Local == "DIV8" {
				inSection = false
				sections = append(sections, domain.Section{
					ID:      sectionID,
					Section: sectionID, // Simplified
					Text:    currentText.String(),
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
