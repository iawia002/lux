package qq

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/iawia002/annie/extractors/types"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type qqVideoInfo struct {
	Fl struct {
		Fi []struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Cname string `json:"cname"`
			Fs    int    `json:"fs"`
		} `json:"fi"`
	} `json:"fl"`
	Vl struct {
		Vi []struct {
			Fn    string `json:"fn"`
			Ti    string `json:"ti"`
			Fvkey string `json:"fvkey"`
			Cl    struct {
				Fc int `json:"fc"`
				Ci []struct {
					Idx int `json:"idx"`
				} `json:"ci"`
			} `json:"cl"`
			Ul struct {
				UI []struct {
					URL string `json:"url"`
				} `json:"ui"`
			} `json:"ul"`
		} `json:"vi"`
	} `json:"vl"`
	Msg string `json:"msg"`
}

type qqKeyInfo struct {
	Key string `json:"key"`
}

const qqPlayerVersion string = "3.2.19.333"

func getVinfo(vid, defn, refer string) (qqVideoInfo, error) {
	html, err := request.Get(
		fmt.Sprintf(
			"http://vv.video.qq.com/getinfo?otype=json&platform=11&defnpayver=1&appver=%s&defn=%s&vid=%s",
			qqPlayerVersion, defn, vid,
		), refer, nil,
	)
	if err != nil {
		return qqVideoInfo{}, err
	}
	jsonStrings := utils.MatchOneOf(html, `QZOutputJson=(.+);$`)
	if jsonStrings == nil || len(jsonStrings) < 2 {
		return qqVideoInfo{}, types.ErrURLParseFailed
	}
	jsonString := jsonStrings[1]
	var data qqVideoInfo
	if err = json.Unmarshal([]byte(jsonString), &data); err != nil {
		return qqVideoInfo{}, err
	}
	return data, nil
}

func genStreams(vid, cdn string, data qqVideoInfo) (map[string]*types.Stream, error) {
	streams := make(map[string]*types.Stream)
	var vkey string
	// number of fragments
	var clips int

	for _, fi := range data.Fl.Fi {
		var fmtIDPrefix string
		var fns []string
		if utils.ItemInSlice(fi.Name, []string{"shd", "fhd"}) {
			fmtIDPrefix = "p"
			fmtIDName := fmt.Sprintf("%s%d", fmtIDPrefix, fi.ID%10000)
			fns = []string{strings.Split(data.Vl.Vi[0].Fn, ".")[0], fmtIDName, "mp4"}
			if len(fns) > 3 {
				// delete ID part
				// e0765r4mwcr.2.mp4 -> e0765r4mwcr.mp4
				fns = append(fns[:1], fns[2:]...)
			}
			clips = data.Vl.Vi[0].Cl.Fc
			if clips == 0 {
				clips = 1
			}
		} else {
			tmpData, err := getVinfo(vid, fi.Name, cdn)
			if err != nil {
				return nil, err
			}
			fns = strings.Split(tmpData.Vl.Vi[0].Fn, ".")
			if len(fns) >= 3 && utils.MatchOneOf(fns[1], `^p(\d{3})$`) != nil {
				fmtIDPrefix = "p"
			}
			clips = tmpData.Vl.Vi[0].Cl.Fc
			if clips == 0 {
				clips = 1
			}
		}

		var urls []*types.Part
		var totalSize int64
		var filename string
		for part := 1; part < clips+1; part++ {
			// Multiple fragments per streams
			if fmtIDPrefix == "p" {
				if len(fns) < 4 {
					// If the number of fragments > 0, the filename needs to add the number of fragments
					// n0687peq62x.p709.mp4 -> n0687peq62x.p709.1.mp4
					fns = append(fns[:2], append([]string{strconv.Itoa(part)}, fns[2:]...)...)
				} else {
					fns[2] = strconv.Itoa(part)
				}
			}
			filename = strings.Join(fns, ".")
			html, err := request.Get(
				fmt.Sprintf(
					"http://vv.video.qq.com/getkey?otype=json&platform=11&appver=%s&filename=%s&format=%d&vid=%s",
					qqPlayerVersion, filename, fi.ID, vid,
				), "", nil,
			)
			if err != nil {
				return nil, err
			}
			jsonStrings := utils.MatchOneOf(html, `QZOutputJson=(.+);$`)
			if jsonStrings == nil || len(jsonStrings) < 2 {
				return nil, types.ErrURLParseFailed
			}
			jsonString := jsonStrings[1]

			var keyData qqKeyInfo
			if err = json.Unmarshal([]byte(jsonString), &keyData); err != nil {
				return nil, err
			}

			vkey = keyData.Key
			if vkey == "" {
				vkey = data.Vl.Vi[0].Fvkey
			}
			realURL := fmt.Sprintf("%s%s?vkey=%s", cdn, filename, vkey)
			size, err := request.Size(realURL, cdn)
			if err != nil {
				return nil, err
			}
			urlData := &types.Part{
				URL:  realURL,
				Size: size,
				Ext:  "mp4",
			}
			urls = append(urls, urlData)
			totalSize += size
		}
		streams[fi.Name] = &types.Stream{
			Parts:   urls,
			Size:    totalSize,
			Quality: fi.Cname,
		}
	}
	return streams, nil
}

type extractor struct{}

// New returns a youtube extractor.
func New() types.Extractor {
	return &extractor{}
}

// Extract is the main function to extract the data.
func (e *extractor) Extract(url string, option types.Options) ([]*types.Data, error) {
	vids := utils.MatchOneOf(url, `vid=(\w+)`, `/(\w+)\.html`)
	if vids == nil || len(vids) < 2 {
		return nil, types.ErrURLParseFailed
	}
	vid := vids[1]

	if len(vid) != 11 {
		u, err := request.Get(url, url, nil)
		if err != nil {
			return nil, err
		}

		vids = utils.MatchOneOf(
			u, `vid=(\w+)`, `vid:\s*["'](\w+)`, `vid\s*=\s*["']\s*(\w+)`,
		)
		if vids == nil || len(vids) < 2 {
			return nil, types.ErrURLParseFailed
		}
		vid = vids[1]
	}

	data, err := getVinfo(vid, "shd", url)
	if err != nil {
		return nil, err
	}

	// API request error
	if data.Msg != "" {
		return nil, errors.New(data.Msg)
	}
	cdn := data.Vl.Vi[0].Ul.UI[0].URL
	streams, err := genStreams(vid, cdn, data)
	if err != nil {
		return nil, err
	}

	return []*types.Data{
		{
			Site:    "腾讯视频 v.qq.com",
			Title:   data.Vl.Vi[0].Ti,
			Type:    types.DataTypeVideo,
			Streams: streams,
			URL:     url,
		},
	}, nil
}
