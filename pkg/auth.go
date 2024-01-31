package pkg

import (
	"net/http"
	"os"
)

func addTifAuthHeaders(request *http.Request) {
	apiKey := os.Getenv("API_KEY")
	token := os.Getenv("TOKEN")
	request.Header.Set("x-api-key", "fuit-pie")
	request.Header.Set("token", token)
	request.Header.Set("Ocp-Apim-Subscription-Key", apiKey)
}
