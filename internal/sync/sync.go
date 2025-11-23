package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"gopkg.in/yaml.v3"
)

var GitHubAPIURL = "https://api.github.com"

type GistFile struct {
	Content string `json:"content"`
}

type Gist struct {
	Files map[string]GistFile `json:"files"`
	Description string        `json:"description"`
	Public      bool          `json:"public"`
}

type Client struct {
	Token      string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		Token: token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetGist(gistID string) (*config.Config, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/gists/%s", GitHubAPIURL, gistID), nil)
	if err != nil {
		return nil, err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "token "+c.Token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get gist: %s", resp.Status)
	}

	var gist Gist
	if err := json.NewDecoder(resp.Body).Decode(&gist); err != nil {
		return nil, err
	}

	file, ok := gist.Files["config.yml"]
	if !ok {
		return nil, fmt.Errorf("config.yml not found in gist")
	}

	var cfg config.Config
	if err := yaml.Unmarshal([]byte(file.Content), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config from gist: %w", err)
	}

	return &cfg, nil
}

func (c *Client) UpdateGist(gistID string, cfg *config.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	gist := Gist{
		Files: map[string]GistFile{
			"config.yml": {Content: string(data)},
		},
		Description: "Entry Configuration (Updated " + time.Now().Format(time.RFC3339) + ")",
	}

	body, err := json.Marshal(gist)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/gists/%s", GitHubAPIURL, gistID), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "token "+c.Token)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update gist: %s - %s", resp.Status, string(bodyBytes))
	}

	return nil
}

func (c *Client) CreateGist(cfg *config.Config, public bool) (string, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}

	gist := Gist{
		Files: map[string]GistFile{
			"config.yml": {Content: string(data)},
		},
		Description: "Entry Configuration",
		Public:      public,
	}

	body, err := json.Marshal(gist)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/gists", GitHubAPIURL), bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "token "+c.Token)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create gist: %s - %s", resp.Status, string(bodyBytes))
	}

	var respData struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", err
	}

	return respData.ID, nil
}
