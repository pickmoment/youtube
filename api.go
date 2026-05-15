package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
)

var (
	reVideoID  = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)
	reVideoURL = regexp.MustCompile(`(?:v=|youtu\.be/)([A-Za-z0-9_-]{11})`)
	reChanID   = regexp.MustCompile(`^UC[A-Za-z0-9_-]{22}$`)
)

const ytBase = "https://www.youtube.com"

type Video struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Channel    string `json:"channel,omitempty"`
	Duration   string `json:"duration,omitempty"`
	UploadDate string `json:"upload_date,omitempty"`
	ViewCount  string `json:"view_count,omitempty"`
	URL        string `json:"url"`
}

type VideoInfo struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Channel     string   `json:"channel"`
	ChannelID   string   `json:"channel_id"`
	ChannelURL  string   `json:"channel_url,omitempty"`
	Duration    string   `json:"duration"`
	PublishDate string   `json:"publish_date,omitempty"`
	UploadDate  string   `json:"upload_date,omitempty"`
	ViewCount   string   `json:"view_count,omitempty"`
	IsLive      bool     `json:"is_live,omitempty"`
	IsPrivate   bool     `json:"is_private,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Thumbnail   string   `json:"thumbnail,omitempty"`
	URL         string   `json:"url"`
}

type Channel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Handle      string `json:"handle,omitempty"`
	Description string `json:"description,omitempty"`
	Subscribers string `json:"subscribers,omitempty"`
	VideoCount  string `json:"video_count,omitempty"`
	URL         string `json:"url"`
}

type TranscriptSegment struct {
	Start    float64 `json:"start"`
	Duration float64 `json:"duration"`
	Text     string  `json:"text"`
}

type Transcript struct {
	VideoID  string              `json:"video_id"`
	URL      string              `json:"url"`
	Lang     string              `json:"lang"`
	Segments []TranscriptSegment `json:"transcript"`
	Text     string              `json:"text"`
}

// JSON navigation helpers

func jget(obj any, keys ...string) any {
	for _, k := range keys {
		m, ok := obj.(map[string]any)
		if !ok {
			return nil
		}
		obj = m[k]
	}
	return obj
}

func jstr(obj any, keys ...string) string {
	s, _ := jget(obj, keys...).(string)
	return s
}

func jbool(obj any, keys ...string) bool {
	b, _ := jget(obj, keys...).(bool)
	return b
}

func jarr(obj any, keys ...string) []any {
	a, _ := jget(obj, keys...).([]any)
	return a
}

func runsText(obj any) string {
	runs := jarr(obj, "runs")
	var sb strings.Builder
	for _, r := range runs {
		sb.WriteString(jstr(r, "text"))
	}
	return sb.String()
}

// extractJSONVar finds a JS variable assignment and decodes the JSON value.
// Uses json.Decoder so it stops cleanly at the end of the JSON object.
func extractJSONVar(html, varName string) (map[string]any, error) {
	marker := varName + " = "
	idx := strings.Index(html, marker)
	if idx < 0 {
		return nil, fmt.Errorf("%s를 페이지에서 찾을 수 없습니다", varName)
	}
	r := strings.NewReader(html[idx+len(marker):])
	var result map[string]any
	if err := json.NewDecoder(r).Decode(&result); err != nil {
		return nil, fmt.Errorf("%s 파싱 오류: %w", varName, err)
	}
	return result, nil
}

func videoIDFromInput(input string) (string, error) {
	if reVideoID.MatchString(input) {
		return input, nil
	}
	if m := reVideoURL.FindStringSubmatch(input); m != nil {
		return m[1], nil
	}
	return "", fmt.Errorf("유효한 YouTube 영상 ID 또는 URL이 아닙니다: %s", input)
}

func channelPageURL(channel string) string {
	return channelBaseURL(channel) + "/videos"
}

func channelBaseURL(channel string) string {
	if strings.HasPrefix(channel, "http") {
		base := strings.TrimRight(channel, "/")
		base = strings.TrimSuffix(base, "/videos")
		return strings.TrimRight(base, "/")
	}
	if reChanID.MatchString(channel) {
		return ytBase + "/channel/" + channel
	}
	handle := channel
	if !strings.HasPrefix(handle, "@") {
		handle = "@" + handle
	}
	return ytBase + "/" + handle
}

func channelInfo(channel string) (*Channel, error) {
	url := channelBaseURL(channel)
	html, err := getHTML(url)
	if err != nil {
		return nil, err
	}

	data, err := extractJSONVar(html, "ytInitialData")
	if err != nil {
		return nil, err
	}

	ch := &Channel{URL: url}

	if meta := jget(data, "metadata", "channelMetadataRenderer"); meta != nil {
		ch.ID = jstr(meta, "externalId")
		ch.Name = jstr(meta, "title")
		ch.Description = jstr(meta, "description")
		if chanURL := jstr(meta, "channelUrl"); chanURL != "" {
			ch.URL = chanURL
		}
	}

	// Old header format
	if h := jget(data, "header", "c4TabbedHeaderRenderer"); h != nil {
		if ch.Name == "" {
			ch.Name = jstr(h, "title")
		}
		ch.Subscribers = jstr(h, "subscriberCountText", "simpleText")
		if ch.Subscribers == "" {
			ch.Subscribers = runsText(jget(h, "subscriberCountText"))
		}
		ch.VideoCount = runsText(jget(h, "videosCountText"))
		ch.Handle = runsText(jget(h, "channelHandleText"))
	}

	// Newer pageHeaderRenderer format
	if ch.Subscribers == "" {
		if h := jget(data, "header", "pageHeaderRenderer"); h != nil {
			vm := jget(h, "content", "pageHeaderViewModel")
			if vm != nil {
				if ch.Name == "" {
					ch.Name = jstr(vm, "title", "dynamicText", "text", "content")
					if ch.Name == "" {
						ch.Name = jstr(vm, "title", "content")
					}
				}
				rows := jarr(jget(vm, "metadata", "contentMetadataViewModel"), "metadataRows")
				for _, row := range rows {
					for _, part := range jarr(row, "metadataParts") {
						text := jstr(part, "text", "content")
						switch {
						case strings.HasPrefix(text, "@") && ch.Handle == "":
							ch.Handle = text
						case strings.Contains(text, "구독자") || strings.Contains(text, "subscriber"):
							ch.Subscribers = text
						case strings.Contains(text, "동영상") || strings.Contains(text, "video"):
							ch.VideoCount = text
						}
					}
				}
			}
		}
	}

	if ch.Name == "" {
		return nil, fmt.Errorf("채널 정보를 찾을 수 없습니다")
	}
	return ch, nil
}

func parseVideoRenderer(vr any) *Video {
	id := jstr(vr, "videoId")
	if id == "" {
		return nil
	}
	title := runsText(jget(vr, "title"))
	channel := runsText(jget(vr, "shortBylineText"))
	if channel == "" {
		channel = runsText(jget(vr, "longBylineText"))
	}
	duration := jstr(vr, "lengthText", "simpleText")
	uploadDate := jstr(vr, "publishedTimeText", "simpleText")

	viewCount := jstr(vr, "viewCountText", "simpleText")
	if viewCount == "" {
		if runs := jarr(jget(vr, "viewCountText"), "runs"); len(runs) > 0 {
			var parts []string
			for _, r := range runs {
				parts = append(parts, jstr(r, "text"))
			}
			viewCount = strings.TrimSpace(strings.Join(parts, ""))
		}
	}

	return &Video{
		ID:         id,
		Title:      title,
		Channel:    channel,
		Duration:   duration,
		UploadDate: uploadDate,
		ViewCount:  viewCount,
		URL:        ytBase + "/watch?v=" + id,
	}
}

func search(query string, limit int) ([]Video, error) {
	resp, err := postInnerTube("search", map[string]any{
		"query": query,
	})
	if err != nil {
		return nil, err
	}

	sections := jarr(jget(resp,
		"contents",
		"twoColumnSearchResultsRenderer",
		"primaryContents",
		"sectionListRenderer",
	), "contents")

	var results []Video
	for _, section := range sections {
		isr := jget(section, "itemSectionRenderer")
		if isr == nil {
			continue
		}
		for _, item := range jarr(isr, "contents") {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			vr, ok := m["videoRenderer"]
			if !ok {
				continue
			}
			vid := parseVideoRenderer(vr)
			if vid != nil {
				results = append(results, *vid)
				if len(results) >= limit {
					return results, nil
				}
			}
		}
	}
	return results, nil
}

func channelVideos(channel string, limit int) ([]Video, error) {
	url := channelPageURL(channel)
	html, err := getHTML(url)
	if err != nil {
		return nil, err
	}

	data, err := extractJSONVar(html, "ytInitialData")
	if err != nil {
		return nil, err
	}

	tabs := jarr(jget(data, "contents", "twoColumnBrowseResultsRenderer"), "tabs")
	var gridContents []any
	for _, tab := range tabs {
		tabRenderer := jget(tab, "tabRenderer")
		if tabRenderer == nil {
			continue
		}
		gc := jarr(jget(tabRenderer, "content", "richGridRenderer"), "contents")
		if len(gc) > 0 {
			gridContents = gc
			break
		}
	}
	if gridContents == nil {
		return nil, fmt.Errorf("채널 영상 목록을 찾을 수 없습니다")
	}

	var results []Video
	for _, item := range gridContents {
		content := jget(item, "richItemRenderer", "content")
		if content == nil {
			continue
		}
		var vid *Video
		if vr := jget(content, "videoRenderer"); vr != nil {
			vid = parseVideoRenderer(vr)
		} else if lvm := jget(content, "lockupViewModel"); lvm != nil {
			vid = parseLockupViewModel(lvm)
		}
		if vid != nil {
			results = append(results, *vid)
			if len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}

func parseLockupViewModel(lvm any) *Video {
	id := jstr(lvm, "contentId")
	if id == "" {
		return nil
	}

	meta := jget(lvm, "metadata", "lockupMetadataViewModel")
	title := jstr(meta, "title", "content")

	var viewCount, uploadDate string
	for _, row := range jarr(jget(meta, "metadata", "contentMetadataViewModel"), "metadataRows") {
		for _, part := range jarr(row, "metadataParts") {
			text := jstr(part, "text", "content")
			switch {
			case strings.Contains(text, "조회수") || strings.Contains(text, "views"):
				viewCount = text
			case viewCount != "" && uploadDate == "":
				uploadDate = text
			}
		}
	}

	var duration string
	for _, overlay := range jarr(jget(lvm, "contentImage", "thumbnailViewModel"), "overlays") {
		for _, badge := range jarr(jget(overlay, "thumbnailBottomOverlayViewModel"), "badges") {
			if t := jstr(badge, "thumbnailBadgeViewModel", "text"); t != "" {
				duration = t
				break
			}
		}
		if duration != "" {
			break
		}
	}

	return &Video{
		ID:         id,
		Title:      title,
		Duration:   duration,
		UploadDate: uploadDate,
		ViewCount:  viewCount,
		URL:        ytBase + "/watch?v=" + id,
	}
}

func videoInfo(video string) (*VideoInfo, error) {
	vid, err := videoIDFromInput(video)
	if err != nil {
		return nil, err
	}

	html, err := getHTML(ytBase + "/watch?v=" + vid)
	if err != nil {
		return nil, err
	}
	m := reInnertubeKey.FindStringSubmatch(html)
	if m == nil {
		return nil, fmt.Errorf("InnerTube API 키를 찾을 수 없습니다")
	}

	playerResp, err := postRaw(innertubeBase+"/player?key="+m[1]+"&prettyPrint=false", map[string]any{
		"context": map[string]any{
			"client": map[string]any{
				"clientName":    "ANDROID",
				"clientVersion": "20.10.38",
			},
		},
		"videoId": vid,
	})
	if err != nil {
		return nil, err
	}

	details := jget(playerResp, "videoDetails")
	if details == nil {
		return nil, fmt.Errorf("영상 정보를 찾을 수 없습니다")
	}

	info := &VideoInfo{
		ID:        jstr(details, "videoId"),
		Title:     jstr(details, "title"),
		Channel:   jstr(details, "author"),
		ChannelID: jstr(details, "channelId"),
		IsLive:    jbool(details, "isLiveContent"),
		IsPrivate: jbool(details, "isPrivate"),
		URL:       ytBase + "/watch?v=" + vid,
	}
	if info.ChannelID != "" {
		info.ChannelURL = ytBase + "/channel/" + info.ChannelID
	}
	if secs := jstr(details, "lengthSeconds"); secs != "" {
		info.Duration = formatDuration(secs)
	}
	info.ViewCount = jstr(details, "viewCount")
	info.Description = jstr(details, "shortDescription")

	for _, kw := range jarr(details, "keywords") {
		if s, ok := kw.(string); ok {
			info.Tags = append(info.Tags, s)
		}
	}

	thumbs := jarr(jget(details, "thumbnail"), "thumbnails")
	if len(thumbs) > 0 {
		info.Thumbnail = jstr(thumbs[len(thumbs)-1], "url")
	}

	if mf := jget(playerResp, "microformat", "playerMicroformatRenderer"); mf != nil {
		if pd := jstr(mf, "publishDate"); pd != "" {
			info.PublishDate = pd
		}
	}

	// ytInitialData: publish date fallback + relative upload date
	if data, err2 := extractJSONVar(html, "ytInitialData"); err2 == nil {
		contents := jarr(jget(data, "contents", "twoColumnWatchNextResults", "results", "results"), "contents")
		for _, c := range contents {
			if pir := jget(c, "videoPrimaryInfoRenderer"); pir != nil {
				if info.PublishDate == "" {
					info.PublishDate = jstr(pir, "dateText", "simpleText")
				}
				info.UploadDate = jstr(pir, "relativeDateText", "simpleText")
				break
			}
		}
	}

	return info, nil
}

func formatDuration(secs string) string {
	n := 0
	for _, c := range secs {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	h, m, s := n/3600, (n%3600)/60, n%60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

// reInnertubeKey extracts the InnerTube API key from the video page HTML.
var reInnertubeKey = regexp.MustCompile(`"INNERTUBE_API_KEY":\s*"([a-zA-Z0-9_-]+)"`)

// xmlTranscript is used for parsing the timedtext XML response.
type xmlTranscript struct {
	Entries []xmlText `xml:"text"`
}
type xmlText struct {
	Start float64 `xml:"start,attr"`
	Dur   float64 `xml:"dur,attr"`
	Text  string  `xml:",chardata"`
}

func transcript(video string, langs []string) (*Transcript, error) {
	vid, err := videoIDFromInput(video)
	if err != nil {
		return nil, err
	}

	// Step 1: Fetch video page and extract InnerTube API key
	html, err := getHTML(ytBase + "/watch?v=" + vid)
	if err != nil {
		return nil, err
	}
	m := reInnertubeKey.FindStringSubmatch(html)
	if m == nil {
		return nil, fmt.Errorf("InnerTube API 키를 찾을 수 없습니다")
	}
	apiKey := m[1]

	// Step 2: Call the player InnerTube endpoint with ANDROID client.
	// The ANDROID client returns timedtext URLs without the exp=xpe PO-token gate.
	playerURL := innertubeBase + "/player?key=" + apiKey + "&prettyPrint=false"
	playerResp, err := postRaw(playerURL, map[string]any{
		"context": map[string]any{
			"client": map[string]any{
				"clientName":    "ANDROID",
				"clientVersion": "20.10.38",
			},
		},
		"videoId": vid,
	})
	if err != nil {
		return nil, err
	}

	tracks := jarr(jget(playerResp, "captions", "playerCaptionsTracklistRenderer"), "captionTracks")
	if len(tracks) == 0 {
		return nil, fmt.Errorf("자막을 찾을 수 없습니다 (video: %s)", vid)
	}

	track := selectTrack(tracks, langs)
	if track == nil {
		return nil, fmt.Errorf("요청한 언어의 자막을 찾을 수 없습니다")
	}

	// Remove fmt=srv3 if present (library does the same), then fetch XML
	captionURL := strings.Replace(jstr(track, "baseUrl"), "&fmt=srv3", "", 1)
	langCode := jstr(track, "languageCode")

	b, err := getBytes(captionURL)
	if err != nil {
		return nil, err
	}

	// Step 3: Parse XML timedtext
	var raw xmlTranscript
	if err := xml.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("자막 파싱 오류: %w", err)
	}

	var segments []TranscriptSegment
	var texts []string
	for _, entry := range raw.Entries {
		text := strings.TrimSpace(unescapeHTML(entry.Text))
		if text == "" {
			continue
		}
		segments = append(segments, TranscriptSegment{
			Start:    roundSec(entry.Start),
			Duration: roundSec(entry.Dur),
			Text:     text,
		})
		texts = append(texts, text)
	}

	return &Transcript{
		VideoID:  vid,
		URL:      ytBase + "/watch?v=" + vid,
		Lang:     langCode,
		Segments: segments,
		Text:     strings.Join(texts, " "),
	}, nil
}

func roundSec(f float64) float64 {
	// round to 3 decimal places
	return float64(int(f*1000+0.5)) / 1000
}

var reHTMLEntity = regexp.MustCompile(`&#(\d+);|&amp;|&lt;|&gt;|&apos;|&quot;`)

func unescapeHTML(s string) string {
	return reHTMLEntity.ReplaceAllStringFunc(s, func(match string) string {
		switch match {
		case "&amp;":
			return "&"
		case "&lt;":
			return "<"
		case "&gt;":
			return ">"
		case "&apos;":
			return "'"
		case "&quot;":
			return `"`
		default:
			// &#NNN;
			if len(match) > 3 {
				n := 0
				for _, c := range match[2 : len(match)-1] {
					n = n*10 + int(c-'0')
				}
				return string(rune(n))
			}
			return match
		}
	})
}

func selectTrack(tracks []any, langs []string) any {
	if len(langs) == 0 {
		// Prefer manually created captions over auto-generated (asr)
		for _, t := range tracks {
			if jstr(t, "kind") != "asr" {
				return t
			}
		}
		if len(tracks) > 0 {
			return tracks[0]
		}
		return nil
	}
	for _, lang := range langs {
		for _, t := range tracks {
			if jstr(t, "languageCode") == lang && jstr(t, "kind") != "asr" {
				return t
			}
		}
		for _, t := range tracks {
			if jstr(t, "languageCode") == lang {
				return t
			}
		}
	}
	return nil
}
