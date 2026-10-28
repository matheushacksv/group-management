package groups

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents the http client
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// NewClient creates a http client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{baseURL, apiKey, &http.Client{
		Timeout: 30 * time.Second,
	}}
}

// ListGroups list all groups for a session
func (c *Client) ListGroups(sessionID string) ([]Group, error) {
	var groups []Group
	path := fmt.Sprintf("/api/sessions/%s/groups", sessionID)

	req, err := http.NewRequest(
		http.MethodGet, c.baseURL+path, nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return []Group{}, fmt.Errorf("error: expected status 200, got %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&groups); err != nil {
		return nil, err
	}

	return groups, nil
}

// GetGroup get a specific WA group for a session
func (c *Client) GetGroup(sessionID, groupID string) (Group, error) {
	var group Group
	path := fmt.Sprintf("/api/sessions/%s/groups/%s", sessionID, groupID)

	req, err := http.NewRequest(
		http.MethodGet, c.baseURL+path, nil,
	)
	if err != nil {
		return Group{}, err
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return Group{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Group{}, fmt.Errorf("error: expected status 200, got %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&group); err != nil {
		return Group{}, err
	}

	return group, nil
}

// ListContacts list all contacts known by a session
func (c *Client) ListContacts(sessionID string) ([]Contact, error) {
	var contacts []Contact
	path := fmt.Sprintf("/api/sessions/%s/contacts", sessionID)

	req, err := http.NewRequest(
		http.MethodGet, c.baseURL+path, nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", c.apiKey)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: expected status 200, got %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&contacts); err != nil {
		return nil, err
	}

	return contacts, nil
}

// RemoveParticipants remove a participant from a group
func (c *Client) RemoveParticipants(sessionID, groupID string, numbers []string) ([]RemoveResult, error) {
	body, err := json.Marshal(map[string][]string{"participants": numbers})
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/api/sessions/%s/groups/%s/participants", sessionID, groupID)

	req, err := http.NewRequest(
		http.MethodDelete,
		c.baseURL+path,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return nil, fmt.Errorf("openwa remove: %d %s", res.StatusCode, msg)
	}

	var out struct {
		Results []RemoveResult `json:"results"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Results, nil
}
