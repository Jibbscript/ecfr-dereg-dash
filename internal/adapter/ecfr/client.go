package ecfr

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/xai/ecfr-dereg-dashboard/internal/domain"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: "https://www.ecfr.gov/api/renderer/v1",
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) GetTitles() ([]domain.Title, error) {
	resp, err := c.client.Get(c.baseURL + "/titles")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var titles []domain.Title
	err = json.NewDecoder(resp.Body).Decode(&titles)
	return titles, err
}

func (c *Client) GetSectionsForTitle(title string) ([]domain.RawSection, error) {
	resp, err := c.client.Get(c.baseURL + "/content/enhanced/title-" + title)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data struct{ Structure []domain.RawSection }
	err = json.NewDecoder(resp.Body).Decode(&data)
	return data.Structure, err
}
