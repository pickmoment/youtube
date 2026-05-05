package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

func toJSON(v any) string {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
	return strings.TrimRight(buf.String(), "\n")
}

func tg(text string) string {
	return "```\n" + text + "\n```"
}

func fmtVideos(data []Video, format string) string {
	if format == "json" {
		return toJSON(data)
	}
	if len(data) == 0 {
		return tg("영상이 없습니다.")
	}
	var lines []string
	for i, v := range data {
		lines = append(lines, fmt.Sprintf("%3d. %s", i+1, v.Title))
		var meta []string
		if v.Channel != "" {
			meta = append(meta, v.Channel)
		}
		if v.Duration != "" {
			meta = append(meta, v.Duration)
		}
		if v.UploadDate != "" {
			meta = append(meta, v.UploadDate)
		}
		if v.ViewCount != "" {
			meta = append(meta, v.ViewCount)
		}
		if len(meta) > 0 {
			lines = append(lines, "     "+strings.Join(meta, "  "))
		}
		lines = append(lines, "     "+v.URL)
	}
	return tg(strings.Join(lines, "\n"))
}

func fmtVideoInfo(v *VideoInfo, format string) string {
	if format == "json" {
		return toJSON(v)
	}
	var lines []string
	lines = append(lines, v.Title)
	lines = append(lines, v.Channel+" "+v.ChannelURL)
	var meta []string
	if v.Duration != "" {
		meta = append(meta, v.Duration)
	}
	if v.PublishDate != "" {
		meta = append(meta, v.PublishDate)
	}
	if v.ViewCount != "" {
		meta = append(meta, v.ViewCount)
	}
	if len(meta) > 0 {
		lines = append(lines, strings.Join(meta, "  "))
	}
	if v.Description != "" {
		desc := v.Description
		if len(desc) > 300 {
			desc = desc[:300] + "..."
		}
		lines = append(lines, "", desc)
	}
	if len(v.Tags) > 0 {
		tags := v.Tags
		if len(tags) > 10 {
			tags = tags[:10]
		}
		lines = append(lines, "", "태그: "+strings.Join(tags, ", "))
	}
	lines = append(lines, "", v.URL)
	return tg(strings.Join(lines, "\n"))
}

func fmtChannel(ch *Channel, format string) string {
	if format == "json" {
		return toJSON(ch)
	}
	var lines []string
	lines = append(lines, ch.Name)
	if ch.Handle != "" {
		lines = append(lines, ch.Handle)
	}
	if ch.ID != "" {
		lines = append(lines, "ID: "+ch.ID)
	}
	if ch.Subscribers != "" {
		lines = append(lines, "구독자: "+ch.Subscribers)
	}
	if ch.VideoCount != "" {
		lines = append(lines, "동영상: "+ch.VideoCount)
	}
	if ch.Description != "" {
		desc := ch.Description
		if len(desc) > 300 {
			desc = desc[:300] + "..."
		}
		lines = append(lines, "", desc)
	}
	lines = append(lines, "", ch.URL)
	return tg(strings.Join(lines, "\n"))
}

func fmtTranscript(data *Transcript, format string, textOnly bool) string {
	if textOnly || format == "text" {
		return data.Text
	}
	if format == "json" {
		return toJSON(data)
	}
	lines := []string{fmt.Sprintf("[%s]  lang: %s", data.URL, data.Lang), ""}
	for _, s := range data.Segments {
		ts := int(s.Start)
		h := ts / 3600
		m := (ts % 3600) / 60
		sec := ts % 60
		var stamp string
		if h > 0 {
			stamp = fmt.Sprintf("%d:%02d:%02d", h, m, sec)
		} else {
			stamp = fmt.Sprintf("%02d:%02d", m, sec)
		}
		lines = append(lines, fmt.Sprintf("[%s] %s", stamp, s.Text))
	}
	return tg(strings.Join(lines, "\n"))
}
