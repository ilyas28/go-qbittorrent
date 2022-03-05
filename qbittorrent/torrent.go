package qbittorrent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"
)

type AddTorrentInputTorrentFilePath struct {
	Path string
}

type AddTorrentInput struct {
	TorrentFilePath    AddTorrentInputTorrentFilePath
	SavePath           string
	Cookie             string
	Category           string
	Tags               string
	Skip_checking      string
	Paused             string
	Root_folder        string
	Rename             string
	Up_limit           int
	Dl_limit           int
	RatioLimit         float64
	seedingTimeLimit   int
	autoTMM            bool
	sequentialDownload string
	firstLastPiecePrio string
}

type GetTorrentsInput struct {
	Filter   string
	Category string
	Tag      string
	Sort     string
	Reverse  bool
	Limit    int
	Offset   int
	Hashes   string
}

type GetTorrentsOutput struct {
	Added_on           int     `json:"added_on"`
	Amount_left        int     `json:"amount_left"`
	Auto_tmm           bool    `json:"auto_tmm"`
	Availability       float64 `json:"availability"`
	Category           string  `json:"category"`
	Completed          int     `json:"completed"`
	Completion_on      int     `json:"completion_on"`
	Content_path       string  `json:"content_path"`
	Dl_limit           int     `json:"dl_limit"`
	Dlspeed            int     `json:"dlspeed"`
	Download_path      string  `json:"download_path"`
	Downloaded         int     `json:"downloaded"`
	Downloaded_session int     `json:"downloaded_session"`
	Eta                int     `json:"eta"`
	F_l_piece_prio     bool    `json:"f_l_piece_prio"`
	Force_start        bool    `json:"force_start"`
	Hash               string  `json:"hash"`
	Infohash_v1        string  `json:"infohash_v1"`
	Infohash_v2        string  `json:"infohash_v2"`
	Last_activity      int     `json:"last_activity"`
	Magnet_uri         string  `json:"magnet_uri"`
	Max_ratio          int     `json:"max_ratio"`
	Max_seeding_time   int     `json:"max_seeding_time"`
	Name               string  `json:"name"`
	Num_complete       int     `json:"num_complete"`
	Num_incomplete     int     `json:"num_incomplete"`
	Num_leechs         int     `json:"num_leechs"`
	Num_seeds          int     `json:"num_seeds"`
	Priority           int     `json:"priority"`
	Progress           float64 `json:"progress"`
	Ratio              float64 `json:"ratio"`
	Save_path          string  `json:"save_path"`
	Seeding_time       int     `json:"seeding_time"`
	Seeding_time_limit int     `json:"seeding_time_limit"`
	Seen_complete      int     `json:"seen_complete"`
	Seq_dl             bool    `json:"seq_dl"`
	Size               int     `json:"size"`
	State              string  `json:"state"`
	Super_seeding      bool    `json:"super_seeding"`
	Tags               string  `json:"tags"`
	Time_active        int     `json:"time_active"`
	Total_size         int     `json:"total_size"`
	Tracker            string  `json:"tracker"`
	Trackers_count     int     `json:"trackers_count"`
	Up_limit           int     `json:"up_limit"`
	Uploaded           int     `json:"uploaded"`
	Uploaded_session   int     `json:"uploaded_session"`
	Upspeed            int     `json:"upspeed"`
}

func createURLWithParams(original string, params interface{}) string {
	v := reflect.ValueOf(params)
	if v.NumField() < 0 {
		return original
	}
	typeOf := v.Type()
	isFirst := true
	newUrl := original

	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).IsZero() {
			continue
		}
		if isFirst {
			isFirst = false
			newUrl = newUrl + "?"
		} else {
			newUrl = newUrl + "&"
		}
		paramValue := url.QueryEscape(fmt.Sprintf("%v", v.Field(i).Interface()))
		newUrl = fmt.Sprintf("%s%s=%s", newUrl, strings.ToLower(typeOf.Field(i).Name), paramValue)
	}

	return newUrl
}

func createMultipartFields(writer *multipart.Writer, params interface{}) error {
	v := reflect.ValueOf(params)
	if v.NumField() < 0 {
		return nil
	}
	typeOf := v.Type()

	for i := 0; i < v.NumField(); i++ {
		if !isValidFormType(v.Field(i).Type()) {
			continue
		}
		if v.Field(i).IsZero() {
			continue
		}

		fieldWriter, err := writer.CreateFormField(strings.ToLower(typeOf.Field(i).Name))
		if err != nil {
			return err
		}

		_, err = io.Copy(fieldWriter, strings.NewReader(fmt.Sprintf("%v", v.Field(i).Interface())))
		if err != nil {
			return err
		}

	}

	return nil
}

func isValidFormType(typeOf reflect.Type) bool {
	validTypes := []reflect.Kind{
		reflect.Int, reflect.Float64, reflect.String, reflect.Bool,
	}
	for _, t := range validTypes {
		if typeOf.Kind() == t {
			return true
		}
	}
	return false
}

func (q *QbittorrentClient) GetTorrents(input GetTorrentsInput) ([]GetTorrentsOutput, error) {
	path := createURLWithParams("torrents/info", input)
	body, res, err := q.get(path)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response error status code: %d", res.StatusCode)
	}

	var getTorrentsResponse []GetTorrentsOutput
	err = json.Unmarshal([]byte(body), &getTorrentsResponse)
	if err != nil {
		return nil, fmt.Errorf("could not parse response data in json, %s", err)
	}

	return getTorrentsResponse, nil
}

func (q *QbittorrentClient) AddTorrent(input AddTorrentInput) (err error) {
	client := &http.Client{Timeout: 30 * time.Second}
	path := "torrents/add"
	url := fmt.Sprintf("%s/%s", q.endpoint, path)
	fmt.Println(url)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	var torrentFileReader io.Reader
	var formWriter io.Writer

	torrentFileReader, err = os.Open(input.TorrentFilePath.Path)
	if err != nil {
		return fmt.Errorf("Torrent file does not exist: %s", err)
	}

	formWriter, err = w.CreateFormFile("torrents", input.TorrentFilePath.Path)
	if err != nil {
		return err
	}

	_, err = io.Copy(formWriter, torrentFileReader)
	if err != nil {
		return err
	}

	err = createMultipartFields(w, input)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.AddCookie(&http.Cookie{
		Name:  "SID",
		Value: q.sid,
	})
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return nil
}
