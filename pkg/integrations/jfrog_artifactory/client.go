package jfrog_artifactory

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/superplanehq/superplane/pkg/core"
)

type Client struct {
	BaseURL     string
	AccessToken string
	http        core.HTTPContext
}

func NewClient(httpCtx core.HTTPContext, ctx core.IntegrationContext) (*Client, error) {
	rawURL, err := ctx.GetConfig("url")
	if err != nil {
		return nil, fmt.Errorf("error getting url: %v", err)
	}

	accessToken, err := ctx.GetConfig("accessToken")
	if err != nil {
		return nil, fmt.Errorf("error getting accessToken: %v", err)
	}

	return &Client{
		BaseURL:     strings.TrimRight(string(rawURL), "/"),
		AccessToken: string(accessToken),
		http:        httpCtx,
	}, nil
}

func (c *Client) execRequest(method, requestURL string, body io.Reader, contentType string, allowedStatuses ...int) (*http.Response, []byte, error) {
	req, err := http.NewRequest(method, requestURL, body)
	if err != nil {
		return nil, nil, fmt.Errorf("error building request: %v", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))

	res, err := c.http.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error executing request: %v", err)
	}
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return res, nil, fmt.Errorf("error reading body: %v", err)
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return res, responseBody, nil
	}

	for _, status := range allowedStatuses {
		if res.StatusCode == status {
			return res, responseBody, nil
		}
	}

	return res, nil, fmt.Errorf("request got %d code: %s", res.StatusCode, string(responseBody))
}

func (c *Client) apiURL(path string) string {
	return fmt.Sprintf("%s%s", c.BaseURL, path)
}

// Ping verifies the Artifactory instance is reachable and credentials are valid.
// The ping endpoint returns plain text, so we skip the default JSON accept header.
func (c *Client) Ping() error {
	req, err := http.NewRequest(http.MethodGet, c.apiURL("/api/system/ping"), nil)
	if err != nil {
		return fmt.Errorf("error building request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(res.Body)
	return fmt.Errorf("request got %d code: %s", res.StatusCode, string(body))
}

// Repository represents a JFrog Artifactory repository.
type Repository struct {
	Key         string `json:"key"`
	Description string `json:"description"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	PackageType string `json:"packageType"`
}

// ListRepositories returns all repositories from the Artifactory instance.
func (c *Client) ListRepositories() ([]Repository, error) {
	_, responseBody, err := c.execRequest(http.MethodGet, c.apiURL("/api/repositories"), nil, "")
	if err != nil {
		return nil, err
	}

	var repos []Repository
	if err := json.Unmarshal(responseBody, &repos); err != nil {
		return nil, fmt.Errorf("error parsing repositories response: %v", err)
	}

	return repos, nil
}

// ArtifactInfo represents metadata about an artifact in Artifactory.
type ArtifactInfo struct {
	Repo         string            `json:"repo"`
	Path         string            `json:"path"`
	Created      string            `json:"created"`
	CreatedBy    string            `json:"createdBy"`
	LastModified string            `json:"lastModified"`
	ModifiedBy   string            `json:"modifiedBy"`
	LastUpdated  string            `json:"lastUpdated"`
	DownloadURI  string            `json:"downloadUri"`
	MimeType     string            `json:"mimeType"`
	Size         string            `json:"size"`
	Checksums    *ArtifactChecksum `json:"checksums"`
	URI          string            `json:"uri"`
}

// ArtifactChecksum contains the checksums of an artifact.
type ArtifactChecksum struct {
	SHA1   string `json:"sha1"`
	MD5    string `json:"md5"`
	SHA256 string `json:"sha256"`
}

// GetArtifactInfo returns metadata about an artifact.
func (c *Client) GetArtifactInfo(repoKey, path string) (*ArtifactInfo, error) {
	path = strings.TrimPrefix(path, "/")
	requestURL := c.apiURL(fmt.Sprintf("/api/storage/%s/%s", repoKey, path))
	_, responseBody, err := c.execRequest(http.MethodGet, requestURL, nil, "")
	if err != nil {
		return nil, err
	}

	var info ArtifactInfo
	if err := json.Unmarshal(responseBody, &info); err != nil {
		return nil, fmt.Errorf("error parsing artifact info response: %v", err)
	}

	return &info, nil
}

// DeployResponse represents the response from deploying an artifact.
type DeployResponse struct {
	Repo        string            `json:"repo"`
	Path        string            `json:"path"`
	Created     string            `json:"created"`
	CreatedBy   string            `json:"createdBy"`
	DownloadURI string            `json:"downloadUri"`
	MimeType    string            `json:"mimeType"`
	Size        string            `json:"size"`
	Checksums   *ArtifactChecksum `json:"checksums"`
	URI         string            `json:"uri"`
}

// DeleteArtifact removes an artifact from the specified repository and path.
func (c *Client) DeleteArtifact(repoKey, path string) error {
	path = strings.TrimPrefix(path, "/")
	requestURL := c.apiURL(fmt.Sprintf("/%s/%s", repoKey, path))
	_, _, err := c.execRequest(http.MethodDelete, requestURL, nil, "", http.StatusNoContent)
	return err
}

// DeployArtifact uploads an artifact to the specified repository and path.
func (c *Client) DeployArtifact(repoKey, path string, content io.Reader, contentType string) (*DeployResponse, error) {
	path = strings.TrimPrefix(path, "/")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	requestURL := c.apiURL(fmt.Sprintf("/%s/%s", repoKey, path))
	_, responseBody, err := c.execRequest(http.MethodPut, requestURL, content, contentType, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	var resp DeployResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		return nil, fmt.Errorf("error parsing deploy response: %v", err)
	}

	return &resp, nil
}
