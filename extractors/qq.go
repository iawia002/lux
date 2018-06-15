package extractors

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/iawia002/annie/downloader"
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

func qqGenFormat(vid, cdn string, data qqVideoInfo) map[string]downloader.FormatData {
	format := map[string]downloader.FormatData{}
	var vkey string
	// number of fragments
	clips := data.Vl.Vi[0].Cl.Fc
	fc := clips
	if clips == 0 {
		clips = 1
	}
	fns := strings.Split(data.Vl.Vi[0].Fn, ".")

	if fc > 0 {
		// If the number of fragments > 0, the filename needs to add the number of fragments
		// n0687peq62x.p709.mp4 -> n0687peq62x.p709.1.mp4
		fns = append(fns[:2], append([]string{"1"}, fns[2:]...)...)
	}
	var fmtIDPrefix string
	for _, fi := range data.Fl.Fi {
		// Multiple formats
		if fi.ID > 100000 {
			fmtIDPrefix = "m"
		} else {
			fmtIDPrefix = "p"
		}
		fmtIDName := fmt.Sprintf("%s%d", fmtIDPrefix, fi.ID%10000)
		fns[1] = fmtIDName
		var urls []downloader.URLData
		var totalSize int64
		var filename string
		for part := 1; part < clips+1; part++ {
			// Multiple fragments per format
			if fc > 0 {
				fns[2] = strconv.Itoa(part)
			}
			filename = strings.Join(fns, ".")
			html := request.Get(
				fmt.Sprintf(
					"http://vv.video.qq.com/getkey?otype=json&platform=11&appver=%s&filename=%s&format=%d&vid=%s", qqPlayerVersion, filename, fi.ID, vid,
				), cdn, nil,
			)
			jsonString := utils.MatchOneOf(html, `QZOutputJson=(.+);$`)[1]
			var keyData qqKeyInfo
			json.Unmarshal([]byte(jsonString), &keyData)
			vkey = keyData.Key
			if vkey == "" {
				vkey = data.Vl.Vi[0].Fvkey
			}
			realURL := fmt.Sprintf("%s%s?vkey=%s", cdn, filename, vkey)
			size := request.Size(realURL, cdn)
			urlData := downloader.URLData{
				URL:  realURL,
				Size: size,
				Ext:  "mp4",
			}
			urls = append(urls, urlData)
			totalSize += size
		}
		format[fi.Name] = downloader.FormatData{
			URLs:    urls,
			Size:    totalSize,
			Quality: fi.Cname,
		}
	}
	format["default"] = format[data.Fl.Fi[len(data.Fl.Fi)-1].Name]
	delete(format, data.Fl.Fi[len(data.Fl.Fi)-1].Name)
	return format
}

// QQ download function
func QQ(url string) downloader.VideoData {
	vid := utils.MatchOneOf(url, `vid=(\w+)`, `/(\w+)\.html`)[1]
	if len(vid) != 11 {
		vid = utils.MatchOneOf(
			request.Get(url, url, nil), `vid=(\w+)`, `vid:\s*["'](\w+)`, `vid\s*=\s*["']\s*(\w+)`,
		)[1]
	}
	html := request.Get(
		fmt.Sprintf(
			"http://vv.video.qq.com/getinfo?otype=json&platform=11&defnpayver=1&appver=%s&defn=shd&vid=%s", qqPlayerVersion, vid,
		), url, nil,
	)
	jsonString := utils.MatchOneOf(html, `QZOutputJson=(.+);$`)[1]
	var data qqVideoInfo
	json.Unmarshal([]byte(jsonString), &data)
	// API request error
	if data.Msg != "" {
		log.Fatal(data.Msg)
	}
	cdn := data.Vl.Vi[0].Ul.UI[len(data.Vl.Vi[0].Ul.UI)-1].URL
	format := qqGenFormat(vid, cdn, data)
	extractedData := downloader.VideoData{
		Site:    "腾讯视频 v.qq.com",
		Title:   utils.FileName(data.Vl.Vi[0].Ti),
		Type:    "video",
		Formats: format,
	}
	extractedData.Download(url)
	return extractedData
}
