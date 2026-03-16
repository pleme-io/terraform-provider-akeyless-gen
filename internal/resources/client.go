package resources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// AkeylessClient wraps the API client with auth.
type AkeylessClient struct {
	GatewayURL string
	Token      string
	HTTPClient *http.Client
}

// NewAkeylessClient creates a new API client.
func NewAkeylessClient(gatewayURL, token string) *AkeylessClient {
	return &AkeylessClient{
		GatewayURL: gatewayURL,
		Token:      token,
		HTTPClient: &http.Client{},
	}
}

// APIRequest represents a generic API request body.
type APIRequest struct {
	fields map[string]interface{}
}

// NewAPIRequest creates a new request with the token pre-set.
func (c *AkeylessClient) NewAPIRequest() *APIRequest {
	r := &APIRequest{
		fields: make(map[string]interface{}),
	}
	if c.Token != "" {
		r.fields["token"] = c.Token
	}
	return r
}

// Set sets a field on the request.
func (r *APIRequest) Set(key string, value interface{}) {
	if value != nil {
		r.fields[key] = value
	}
}

// SetString sets a string field.
func (r *APIRequest) SetString(key, value string) {
	if value != "" {
		r.fields[key] = value
	}
}

// SetStringSlice sets a string slice field.
func (r *APIRequest) SetStringSlice(key string, value []string) {
	if len(value) > 0 {
		r.fields[key] = value
	}
}

// SetBool sets a boolean field.
func (r *APIRequest) SetBool(key string, value bool) {
	r.fields[key] = value
}

// SetInt64 sets an int64 field.
func (r *APIRequest) SetInt64(key string, value int64) {
	r.fields[key] = value
}

// MarshalJSON implements json.Marshaler.
func (r *APIRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.fields)
}

// Call makes an API POST request and returns the parsed response.
func (c *AkeylessClient) Call(ctx context.Context, endpoint string, body *APIRequest) (map[string]interface{}, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	url := fmt.Sprintf("%s%s", c.GatewayURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request to %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d on %s: %s", resp.StatusCode, endpoint, string(respBody))
	}

	var result map[string]interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("decoding response from %s: %w", endpoint, err)
		}
	}

	return result, nil
}

// ExpandStringSet converts a TF Set to a Go string slice.
func ExpandStringSet(ctx context.Context, elements []string) []string {
	return elements
}

// GetNestedString retrieves a string from a nested map using dot-notation path.
func GetNestedString(data map[string]interface{}, path string) (string, bool) {
	parts := splitPath(path)
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			if v, ok := current[part]; ok {
				switch val := v.(type) {
				case string:
					return val, true
				case float64:
					return fmt.Sprintf("%v", val), true
				case bool:
					return fmt.Sprintf("%v", val), true
				default:
					return fmt.Sprintf("%v", val), true
				}
			}
			return "", false
		}

		if next, ok := current[part]; ok {
			if m, ok := next.(map[string]interface{}); ok {
				current = m
			} else {
				return "", false
			}
		} else {
			return "", false
		}
	}

	return "", false
}

// GetNestedStringSlice retrieves a string slice from a nested map.
func GetNestedStringSlice(data map[string]interface{}, path string) ([]string, bool) {
	parts := splitPath(path)
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			if v, ok := current[part]; ok {
				if arr, ok := v.([]interface{}); ok {
					result := make([]string, 0, len(arr))
					for _, item := range arr {
						if s, ok := item.(string); ok {
							result = append(result, s)
						}
					}
					return result, true
				}
			}
			return nil, false
		}

		if next, ok := current[part]; ok {
			if m, ok := next.(map[string]interface{}); ok {
				current = m
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	}

	return nil, false
}

func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, c := range path {
		if c == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
