package qbittorrent

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type QbittorrentClient struct {
	endpoint *url.URL
	username string
	password string
	sid      string
}

func NewQbittorrentClient(endpoint, username, password string) (*QbittorrentClient, error) {
	parsedEndpoint, err := url.Parse(endpoint + "/api/v2")
	if err != nil {
		return nil, fmt.Errorf("invalid qbittorrent endpoint, %s", err)
	}

	qbt := &QbittorrentClient{
		endpoint: parsedEndpoint,
		username: username,
		password: password}
	err = qbt.login()
	if err != nil {
		return nil, fmt.Errorf("Could not login to qbittorrent: %s", err)
	}
	return qbt, nil
}

func (q *QbittorrentClient) newRequest(method, path string, params url.Values, payload io.Reader) (request *http.Request, err error) {
	url := fmt.Sprintf("%s/%s", q.endpoint, path)
	request, err = http.NewRequest(method, url, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	if q.sid != "" {
		request.AddCookie(&http.Cookie{
			Name:  "SID",
			Value: q.sid,
		})
	}
	return request, err
}

func (q *QbittorrentClient) do(request *http.Request) (body []byte, resp *http.Response, err error) {
	client := &http.Client{Timeout: 30 * time.Second}
	log.Printf("calling: %s %s", request.Method, request.URL)
	resp, err = client.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	return body, resp, err
}

func (q *QbittorrentClient) get(path string) ([]byte, *http.Response, error) {
	request, err := q.newRequest(http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	return q.do(request)
}

func (q *QbittorrentClient) post(path string, params url.Values) ([]byte, *http.Response, error) {
	request, err := q.newRequest(http.MethodPost, path, params, nil)
	request.URL.RawQuery = params.Encode()
	request.Header.Set("Content-Type", "multipart/form-data")
	if err != nil {
		return nil, nil, err
	}
	return q.do(request)
}

func getCookie(cookies []*http.Cookie, name string) (string, error) {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.Value, nil
		}
	}
	return "", fmt.Errorf("No cookie found with name %s", name)
}

func (q *QbittorrentClient) login() error {
	path := fmt.Sprintf("auth/login?username=%s&password=%s", q.username, q.password)
	_, res, err := q.get(path)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("response error status code: %d", res.StatusCode)
	}
	sid, err := getCookie(res.Cookies(), "SID")
	if err != nil {
		return err
	}
	q.sid = sid
	return nil
}
