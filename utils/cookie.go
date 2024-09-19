package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func ReadCookie(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	if _, fileErr := os.Stat(raw); fileErr == nil {
		data, err := os.ReadFile(raw)
		if err != nil {
			return "", fmt.Errorf("read cookie from file failed, %s", err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	if strings.HasPrefix(raw, "https://") || strings.HasPrefix(raw, "http://") {
		req, _ := http.NewRequest(http.MethodGet, raw, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("read cookie from api failed, err: %s, code: %d", err, resp.StatusCode)
		}
		defer resp.Body.Close()
		txt, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("read cookie from api failed, err: %s", err)
		}
		return string(txt), nil
	}
	return "", nil
}
