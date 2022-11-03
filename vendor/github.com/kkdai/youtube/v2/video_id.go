package youtube

import (
	"regexp"
	"strings"
)

var videoRegexpList = []*regexp.Regexp{
	regexp.MustCompile(`(?:v|embed|shorts|watch\?v)(?:=|/)([^"&?/=%]{11})`),
	regexp.MustCompile(`(?:=|/)([^"&?/=%]{11})`),
	regexp.MustCompile(`([^"&?/=%]{11})`),
}

// ExtractVideoID extracts the videoID from the given string
func ExtractVideoID(videoID string) (string, error) {
	if strings.Contains(videoID, "youtu") || strings.ContainsAny(videoID, "\"?&/<%=") {
		for _, re := range videoRegexpList {
			if isMatch := re.MatchString(videoID); isMatch {
				subs := re.FindStringSubmatch(videoID)
				videoID = subs[1]
			}
		}
	}

	if strings.ContainsAny(videoID, "?&/<%=") {
		return "", ErrInvalidCharactersInVideoID
	}
	if len(videoID) < 10 {
		return "", ErrVideoIDMinLength
	}

	return videoID, nil
}
