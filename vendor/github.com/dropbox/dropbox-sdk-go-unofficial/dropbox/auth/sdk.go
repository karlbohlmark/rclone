package auth

import (
	"encoding/json"
	"mime"
	"net/http"
	"strconv"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
)

type RateLimitAPIError struct {
	dropbox.APIError
	RateLimitError *RateLimitError `json:"error"`
}

// HandleCommonAuthErrors handles common authentication errors
func HandleCommonAuthErrors(c dropbox.Config, resp *http.Response, body []byte) error {
	if resp.StatusCode == http.StatusTooManyRequests {
		var apiError RateLimitAPIError
		// Check content-type
		contentType, _, _ := mime.ParseMediaType(resp.Header.Get("content-type"))
		if contentType == "application/json" {
			if err := json.Unmarshal(body, &apiError); err != nil {
				c.LogDebug("Error unmarshaling '%s' into JSON", body)
				return err
			}
		} else { // assume plain text
			apiError.ErrorSummary = string(body)
			reason := RateLimitReason{dropbox.Tagged{Tag: RateLimitReasonTooManyRequests}}
			apiError.RateLimitError = NewRateLimitError(&reason)
			timeout, _ := strconv.ParseInt(resp.Header.Get("retry-after"), 10, 64)
			apiError.RateLimitError.RetryAfter = uint64(timeout)
		}
		return apiError
	}
	return nil
}
