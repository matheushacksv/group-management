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

// SessionPhone returns the phone number of the session's own WhatsApp account (the bot)
func (c *Client) SessionPhone(sessionID string) (string, error) {
	path := fmt.Sprintf("/api/sessions/%s", sessionID)

	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-API-Key", c.apiKey)

	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error: expected status 200, got %d", res.StatusCode)
	}

	var out struct {
		Phone string `json:"phone"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Phone == "" {
		return "", fmt.Errorf("openwa session %s has no phone, is it connected?", sessionID)
	}
	return out.Phone, nil
}

// LeaveGroup makes the bot leave a group; a group that no longer exists counts as left
func (c *Client) LeaveGroup(sessionID, groupID string) error {
	path := fmt.Sprintf("/api/sessions/%s/groups/%s/leave", sessionID, groupID)

	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", c.apiKey)

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNotFound {
		msg, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return fmt.Errorf("openwa leave: %d %s", res.StatusCode, msg)
	}
	return nil
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
