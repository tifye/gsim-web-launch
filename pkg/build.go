package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Build struct {
	Id      string `json:"id"`
	BlobUrl string `json:"blob"`
}

func FetchLatestRelease(bundleType string) (*Build, error) {
	const baseUrl = "https://hqvrobotics.azure-api.net"
	url := fmt.Sprintf("%s/bundles/indexes/%s?count=1", baseUrl, bundleType)
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

	var builds []Build
	if err = json.Unmarshal(body, &builds); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %v", err)
	}

	if len(builds) == 0 {
		return nil, errors.New("no builds found")
	}

	builds[0].BlobUrl = fmt.Sprintf("%s/bundles/blob/%s", baseUrl, builds[0].BlobUrl)
	return &builds[0], nil
}
