package infrastructure

import (
	"encoding/json"
	"fmt"
	"genkit-examples/internal/book"
	"net/http"
	"net/url"
	"time"
)

type googleBookClient struct {
	httpClient *http.Client
	maxResults int
}

// NewGoogleBookClient creates a Google Books API client.
// apiKey can be empty for public/unauthenticated requests.
// maxResults controls how many results to return (1-40). Values outside this range are clamped.
func NewGoogleBookClient(maxResults int) book.Repository {
	if maxResults < 1 {
		maxResults = 10
	}
	if maxResults > 40 {
		maxResults = 40
	}
	return &googleBookClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		maxResults: maxResults,
	}
}

// Search implements Repository by querying the Google Books Volumes API.
func (c *googleBookClient) Search(query string) ([]book.Book, error) {
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	if c.maxResults == 0 {
		c.maxResults = 10
	}

	const base = "https://www.googleapis.com/books/v1/volumes"
	values := url.Values{}
	values.Set("q", query)
	values.Set("maxResults", fmt.Sprintf("%d", c.maxResults))

	reqURL := fmt.Sprintf("%s?%s", base, values.Encode())
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("google books API error: status %s", resp.Status)
	}

	var apiResp struct {
		Items []struct {
			ID         string `json:"id"`
			VolumeInfo struct {
				Title         string   `json:"title"`
				Description   string   `json:"description"`
				Subtitle      string   `json:"subtitle"`
				Authors       []string `json:"authors"`
				Publisher     string   `json:"publisher"`
				PublishedDate string   `json:"publishedDate"`
			} `json:"volumeInfo"`
			SaleInfo struct {
				BuyLink string `json:"buyLink"`
			} `json:"saleInfo"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	results := make([]book.Book, 0, len(apiResp.Items))
	for _, it := range apiResp.Items {
		results = append(results, book.Book{
			ID:      it.ID,
			Title:   it.VolumeInfo.Title,
			Summary: it.VolumeInfo.Description,
			Metadata: map[string]any{
				"authors":       it.VolumeInfo.Authors,
				"publisher":     it.VolumeInfo.Publisher,
				"publishedDate": it.VolumeInfo.PublishedDate,
				"buyLink":       it.SaleInfo.BuyLink,
			},
		})
	}

	return results, nil
}
