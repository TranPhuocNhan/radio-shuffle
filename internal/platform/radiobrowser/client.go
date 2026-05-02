package radiobrowser

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	defaultBaseURL    = "https://de1.api.radio-browser.info"
	defaultMaxRetries = 3
	defaultTimeout    = 30 * time.Second
)

// apiStation mirrors the exact JSON shape returned by radio-browser.
// It is unexported and only used for deserialization inside this package.
type apiStation struct {
	UUID        string `json:"stationuuid"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	URLResolved string `json:"url_resolved"`
	Homepage    string `json:"homepage"`
	Favicon     string `json:"favicon"`
	Tags        string `json:"tags"`
	Country     string `json:"country"`
	CountryCode string `json:"countrycode"`
	State       string `json:"state"`
	Language    string `json:"language"`
	Codec       string `json:"codec"`
	Bitrate     int    `json:"bitrate"`
	Votes       int    `json:"votes"`
	LastCheckOK int    `json:"lastcheckok"` // API returns 0 or 1
}

// Station is the domain type returned by FetchStations.
// It has no json tags — it is decoupled from the external API field names.
type Station struct {
	UUID        string
	Name        string
	URL         string
	URLResolved string
	Homepage    string
	Favicon     string
	Tags        string
	Country     string
	CountryCode string
	State       string
	Language    string
	Codec       string
	Bitrate     int
	Votes       int
	LastCheckOK bool // converted from API's 0/1 integer
}

// FetchParams controls pagination for FetchStations.
type FetchParams struct {
	Limit  int
	Offset int
}

// Client fetches stations from the radio-browser API.
type Client struct {
	http       *http.Client
	baseURL    string
	maxRetries int
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient replaces the default http.Client.
// Use this to inject a custom transport, timeout, or a test double.
func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) { cl.http = c }
}

// WithMaxRetries sets how many times a transient failure is retried (default 3).
func WithMaxRetries(n int) Option {
	return func(cl *Client) { cl.maxRetries = n }
}

// NewClient constructs a Client. If baseURL is empty the default public endpoint is used.
func NewClient(baseURL string, opts ...Option) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	c := &Client{
		http:       &http.Client{Timeout: defaultTimeout},
		baseURL:    baseURL,
		maxRetries: defaultMaxRetries,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// FetchStations calls /json/stations/search and returns at most params.Limit stations
// starting at params.Offset. Transient errors are retried with exponential backoff.
func (c *Client) FetchStations(ctx context.Context, params FetchParams) ([]Station, error) {
	endpoint, err := url.Parse(c.baseURL + "/json/stations/search")
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	q := endpoint.Query()
	q.Set("limit", strconv.Itoa(params.Limit))
	q.Set("offset", strconv.Itoa(params.Offset))
	q.Set("order", "stationuuid")
	q.Set("hidebroken", "true")
	endpoint.RawQuery = q.Encode()

	var raw []apiStation
	if err := c.doWithRetry(ctx, http.MethodGet, endpoint.String(), &raw); err != nil {
		return nil, err
	}

	stations := make([]Station, len(raw))
	for i, a := range raw {
		stations[i] = toStation(a)
	}
	return stations, nil
}

// doWithRetry executes an HTTP GET and retries on transient failures using
// exponential backoff (1s, 2s, 4s, … capped at 8s).
func (c *Client) doWithRetry(ctx context.Context, method, rawURL string, dest any) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			if backoff > 8*time.Second {
				backoff = 8 * time.Second
			}
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
		if err != nil {
			return fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("User-Agent", "radio-shuffle-syncer/1.0")
		req.Header.Set("Accept", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: %w", attempt+1, err)
			continue
		}

		if isTransient(resp.StatusCode) {
			resp.Body.Close()
			lastErr = fmt.Errorf("attempt %d: radio-browser returned %d", attempt+1, resp.StatusCode)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("radio-browser returned %d", resp.StatusCode)
		}

		decErr := json.NewDecoder(resp.Body).Decode(dest)
		resp.Body.Close()
		if decErr != nil {
			return fmt.Errorf("decode response: %w", decErr)
		}
		return nil
	}

	return fmt.Errorf("all %d attempt(s) failed: %w", c.maxRetries+1, lastErr)
}

// isTransient reports whether a status code is worth retrying.
func isTransient(code int) bool {
	switch code {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return false
}

// toStation maps the internal API struct to the clean domain type.
func toStation(a apiStation) Station {
	return Station{
		UUID:        a.UUID,
		Name:        a.Name,
		URL:         a.URL,
		URLResolved: a.URLResolved,
		Homepage:    a.Homepage,
		Favicon:     a.Favicon,
		Tags:        a.Tags,
		Country:     a.Country,
		CountryCode: a.CountryCode,
		State:       a.State,
		Language:    a.Language,
		Codec:       a.Codec,
		Bitrate:     a.Bitrate,
		Votes:       a.Votes,
		LastCheckOK: a.LastCheckOK == 1,
	}
}
