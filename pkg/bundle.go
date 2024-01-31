package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type BundleType struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
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

func FetchBundleTypes() ([]BundleType, error) {
	const url = "https://hqvrobotics.azure-api.net/bundles/types"
	client := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	addTifAuthHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("response failed with %s", resp.Status)
	}

	if err = resp.Body.Close(); err != nil {
		return nil, fmt.Errorf("error closing response body: %v", err)
	}

	var bundleTypes []BundleType
	if err = json.Unmarshal(body, &bundleTypes); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %v", err)
	}

	return bundleTypes, nil
}
