package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

const (
	userAgent     = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
	innertubeBase = "https://www.youtube.com/youtubei/v1"
	innertubeVer  = "2.20231121.08.00"
	innertubeKey  = "AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"
)

// sessionClient shares cookies across requests (needed for timedtext access)
var sessionClient = newSessionClient()

func newSessionClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Timeout: 20 * time.Second,
		Jar:     jar,
	}
}

var httpClient = &http.Client{Timeout: 20 * time.Second}

func getHTML(url string) (string, error) {
	return getHTMLWith(sessionClient, url)
}

func getHTMLWith(client *http.Client, url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "ko-KR,ko;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,*/*")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}

func getBytes(url string) ([]byte, error) {
	return getBytesWithClient(sessionClient, url)
}

func getBytesWithClient(client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// postRaw posts JSON to an arbitrary URL and returns parsed JSON response.
func postRaw(url string, body map[string]any) (map[string]any, error) {
	rawBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(rawBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")
	resp, err := sessionClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func postInnerTube(endpoint string, body map[string]any) (map[string]any, error) {
	body["context"] = map[string]any{
		"client": map[string]any{
			"hl":            "ko",
			"gl":            "KR",
			"clientName":    "WEB",
			"clientVersion": innertubeVer,
		},
	}
	rawBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	url := innertubeBase + "/" + endpoint + "?key=" + innertubeKey + "&prettyPrint=false"
	req, err := http.NewRequest("POST", url, bytes.NewReader(rawBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-YouTube-Client-Name", "1")
	req.Header.Set("X-YouTube-Client-Version", innertubeVer)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}
