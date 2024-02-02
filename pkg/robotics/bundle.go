package robotics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type BundleRegistry struct {
	baseUrl string
}

type BundleType struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Build struct {
	Id      string `json:"id"`
	BlobUrl string `json:"blob"`
}

func NewBundleRegistry(baseUrl string) *BundleRegistry {
	return &BundleRegistry{
		baseUrl: baseUrl,
	}
}

func (r *BundleRegistry) FetchBundleTypes(ctx context.Context) ([]BundleType, error) {
	url := fmt.Sprintf("%s/bundles/types", r.baseUrl)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	AddTifAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("response failed with %s", resp.Status)
	}

	var bundleTypes []BundleType
	if err = json.Unmarshal(body, &bundleTypes); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %v", err)
	}

	return bundleTypes, nil
}

func (r *BundleRegistry) FetchLatestRelease(ctx context.Context, bundleType string) (*Build, error) {
	url := fmt.Sprintf("%s/bundles/indexes/%s?count=1", r.baseUrl, bundleType)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	AddTifAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("response failed with %s", resp.Status)
	}

	var builds []Build
	if err = json.Unmarshal(body, &builds); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %v", err)
	}

	if len(builds) == 0 {
		return nil, errors.New("no builds found")
	}

	builds[0].BlobUrl = fmt.Sprintf("%s/bundles/blob/%s", r.baseUrl, builds[0].BlobUrl)
	return &builds[0], nil
}

func FilterBundleTypes(types []BundleType, platform Platform) []BundleType {
	var filtered []BundleType
	subStr := "-" + string(platform) + "-Win"
	for _, t := range types {
		if strings.Contains(t.Name, subStr) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}
