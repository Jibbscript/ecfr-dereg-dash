package govinfo

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"cloud.google.com/go/storage"

	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/domain"
)

type Client struct {
	baseURL string
	client  *http.Client

	gcsClient     *storage.Client
	rawBucketName string
	rawPrefix     string
}

func NewClient(ctx context.Context, rawBucketName, rawPrefix string) (*Client, error) {
	gcs, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &Client{
		baseURL:       "https://www.govinfo.gov/bulkdata/json/ECFR",
		client:        &http.Client{Timeout: 10 * time.Minute},
		gcsClient:     gcs,
		rawBucketName: rawBucketName,
		rawPrefix:     rawPrefix,
	}, nil
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

func (c *Client) objectPath(xmlName string) string {
	if c.rawPrefix == "" {
		return xmlName
	}
	return path.Join(c.rawPrefix, xmlName)
}

// DownloadTitleXML downloads the latest XML for a given title into GCS.
func (c *Client) DownloadTitleXML(ctx context.Context, title int) (string, error) {
	// NOTE: The GovInfo Bulk Data JSON API is currently returning 404s.
	// We fallback to using the predictable XML paths directly.
	// Pattern: https://www.govinfo.gov/bulkdata/ECFR/title-{title}/ECFR-title{title}.xml

	xmlName := fmt.Sprintf("ECFR-title%d.xml", title)
	xmlLink := fmt.Sprintf("https://www.govinfo.gov/bulkdata/ECFR/title-%d/%s", title, xmlName)

	// Step 5: Download file to GCS
	objPath := c.objectPath(xmlName)
	obj := c.gcsClient.Bucket(c.rawBucketName).Object(objPath)

	// Check if already exists
	if _, err := obj.Attrs(ctx); err == nil {
		return objPath, nil
	}

	resp, err := c.client.Get(xmlLink)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read body for error details
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))

		// If 404, it might be a missing/reserved title (not all 1-50 exist).
		// We return a specific error that the caller can check to skip gracefully.
		if resp.StatusCode == http.StatusNotFound {
			return "", domain.ErrNotFound
		}

		return "", fmt.Errorf("failed to download XML from %s: status %s, body: %q", xmlLink, resp.Status, string(bodyBytes))
	}

	w := obj.NewWriter(ctx)
	if _, err := io.Copy(w, resp.Body); err != nil {
		_ = w.Close()
		return "", fmt.Errorf("writing XML to GCS object %q: %w", objPath, err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("closing GCS writer for %q: %w", objPath, err)
	}

	return objPath, nil
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

	// Check for 404 or other errors
	if resp.StatusCode != http.StatusOK {
		// Read body for error details
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("GET %s: status %s, body: %q", url, resp.Status, string(bodyBytes))
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

// ParseTitleXML reads XML from GCS instead of local disk.
func (c *Client) ParseTitleXML(ctx context.Context, objPath string) ([]domain.Section, error) {
	rc, err := c.gcsClient.Bucket(c.rawBucketName).Object(objPath).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	decoder := xml.NewDecoder(rc)

	var sections []domain.Section
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
