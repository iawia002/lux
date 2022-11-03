package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	sjson "github.com/bitly/go-simplejson"
)

var (
	playlistIDRegex    = regexp.MustCompile("^[A-Za-z0-9_-]{13,42}$")
	playlistInURLRegex = regexp.MustCompile("[&?]list=([A-Za-z0-9_-]{13,42})(&.*)?$")
)

type Playlist struct {
	ID          string
	Title       string
	Description string
	Author      string
	Videos      []*PlaylistEntry
}

type PlaylistEntry struct {
	ID         string
	Title      string
	Author     string
	Duration   time.Duration
	Thumbnails Thumbnails
}

func extractPlaylistID(url string) (string, error) {
	if playlistIDRegex.Match([]byte(url)) {
		return url, nil
	}

	matches := playlistInURLRegex.FindStringSubmatch(url)

	if matches != nil {
		return matches[1], nil
	}

	return "", ErrInvalidPlaylist
}

// structs for playlist extraction

// Title: metadata.playlistMetadataRenderer.title | sidebar.playlistSidebarRenderer.items[0].playlistSidebarPrimaryInfoRenderer.title.runs[0].text
// Description: metadata.playlistMetadataRenderer.description
// Author: sidebar.playlistSidebarRenderer.items[1].playlistSidebarSecondaryInfoRenderer.videoOwner.videoOwnerRenderer.title.runs[0].text

// Videos: contents.twoColumnBrowseResultsRenderer.tabs[0].tabRenderer.content.sectionListRenderer.contents[0].itemSectionRenderer.contents[0].playlistVideoListRenderer.contents
// ID: .videoId
// Title: title.runs[0].text
// Author: .shortBylineText.runs[0].text
// Duration: .lengthSeconds
// Thumbnails .thumbnails

// TODO?: Author thumbnails: sidebar.playlistSidebarRenderer.items[0].playlistSidebarPrimaryInfoRenderer.thumbnailRenderer.playlistVideoThumbnailRenderer.thumbnail.thumbnails
func (p *Playlist) parsePlaylistInfo(ctx context.Context, client *Client, body []byte) (err error) {
	var j *sjson.Json
	j, err = sjson.NewJson(body)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("JSON parsing error: %v", r)
		}
	}()

	renderer := j.GetPath("alerts").GetIndex(0).GetPath("alertRenderer")
	if renderer != nil && renderer.GetPath("type").MustString() == "ERROR" {
		message := renderer.GetPath("text", "runs").GetIndex(0).GetPath("text").MustString()

		return ErrPlaylistStatus{Reason: message}
	}

	p.Title = j.GetPath("metadata", "playlistMetadataRenderer", "title").MustString()
	p.Description = j.GetPath("metadata", "playlistMetadataRenderer", "description").MustString()
	p.Author = j.GetPath("sidebar", "playlistSidebarRenderer", "items").GetIndex(1).
		GetPath("playlistSidebarSecondaryInfoRenderer", "videoOwner", "videoOwnerRenderer", "title", "runs").
		GetIndex(0).Get("text").MustString()
	vJSON, err := j.GetPath("contents", "twoColumnBrowseResultsRenderer", "tabs").GetIndex(0).
		GetPath("tabRenderer", "content", "sectionListRenderer", "contents").GetIndex(0).
		GetPath("itemSectionRenderer", "contents").GetIndex(0).
		GetPath("playlistVideoListRenderer", "contents").MarshalJSON()

	entries, continuation, err := extractPlaylistEntries(vJSON)
	if err != nil {
		return err
	}

	p.Videos = entries

	for continuation != "" {
		data := prepareInnertubePlaylistData(continuation, true, webClient)

		body, err := client.httpPostBodyBytes(ctx, "https://www.youtube.com/youtubei/v1/browse?key="+webClient.key, data)
		if err != nil {
			return err
		}

		j, err := sjson.NewJson(body)
		if err != nil {
			return err
		}

		vJSON, err := j.GetPath("onResponseReceivedActions").GetIndex(0).
			GetPath("appendContinuationItemsAction", "continuationItems").MarshalJSON()

		if err != nil {
			return err
		}

		entries, token, err := extractPlaylistEntries(vJSON)
		if err != nil {
			return err
		}

		p.Videos, continuation = append(p.Videos, entries...), token
	}

	return err
}

func extractPlaylistEntries(data []byte) ([]*PlaylistEntry, string, error) {
	var vids []*videosJSONExtractor

	if err := json.Unmarshal(data, &vids); err != nil {
		return nil, "", err
	}

	entries := make([]*PlaylistEntry, 0, len(vids))

	var continuation string
	for _, v := range vids {
		if v.Renderer == nil {
			if v.Continuation.Endpoint.Command.Token != "" {
				continuation = v.Continuation.Endpoint.Command.Token
			}

			continue
		}

		entries = append(entries, v.PlaylistEntry())
	}

	return entries, continuation, nil
}

type videosJSONExtractor struct {
	Renderer *struct {
		ID        string   `json:"videoId"`
		Title     withRuns `json:"title"`
		Author    withRuns `json:"shortBylineText"`
		Duration  string   `json:"lengthSeconds"`
		Thumbnail struct {
			Thumbnails []Thumbnail `json:"thumbnails"`
		} `json:"thumbnail"`
	} `json:"playlistVideoRenderer"`
	Continuation struct {
		Endpoint struct {
			Command struct {
				Token string `json:"token"`
			} `json:"continuationCommand"`
		} `json:"continuationEndpoint"`
	} `json:"continuationItemRenderer"`
}

func (vje videosJSONExtractor) PlaylistEntry() *PlaylistEntry {
	ds, err := strconv.Atoi(vje.Renderer.Duration)
	if err != nil {
		panic("invalid video duration: " + vje.Renderer.Duration)
	}
	return &PlaylistEntry{
		ID:         vje.Renderer.ID,
		Title:      vje.Renderer.Title.String(),
		Author:     vje.Renderer.Author.String(),
		Duration:   time.Second * time.Duration(ds),
		Thumbnails: vje.Renderer.Thumbnail.Thumbnails,
	}
}

type withRuns struct {
	Runs []struct {
		Text string `json:"text"`
	} `json:"runs"`
}

func (wr withRuns) String() string {
	if len(wr.Runs) > 0 {
		return wr.Runs[0].Text
	}
	return ""
}
