package youtube

import (
	"fmt"
)

const (
	ErrCipherNotFound             = constError("cipher not found")
	ErrSignatureTimestampNotFound = constError("signature timestamp not found")
	ErrInvalidCharactersInVideoID = constError("invalid characters in video id")
	ErrVideoIDMinLength           = constError("the video id must be at least 10 characters long")
	ErrReadOnClosedResBody        = constError("http: read on closed response body")
	ErrNotPlayableInEmbed         = constError("embedding of this video has been disabled")
	ErrLoginRequired              = constError("login required to confirm your age")
	ErrVideoPrivate               = constError("user restricted access to this video")
	ErrInvalidPlaylist            = constError("no playlist detected or invalid playlist ID")
)

type constError string

func (e constError) Error() string {
	return string(e)
}

type ErrPlayabiltyStatus struct {
	Status string
	Reason string
}

func (err ErrPlayabiltyStatus) Error() string {
	return fmt.Sprintf("cannot playback and download, status: %s, reason: %s", err.Status, err.Reason)
}

// ErrUnexpectedStatusCode is returned on unexpected HTTP status codes
type ErrUnexpectedStatusCode int

func (err ErrUnexpectedStatusCode) Error() string {
	return fmt.Sprintf("unexpected status code: %d", err)
}

type ErrPlaylistStatus struct {
	Reason string
}

func (err ErrPlaylistStatus) Error() string {
	return fmt.Sprintf("could not load playlist: %s", err.Reason)
}
