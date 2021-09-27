package downloader

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/tomwhy/PlaylistMusicDownloader/apis/downloader/model"
)

type Mp3DownloadAPI struct {
	key string
}

func NewMp3DownloadAPI(key string) *Mp3DownloadAPI {
	return &Mp3DownloadAPI{key}
}

func (api *Mp3DownloadAPI) DownloadSong(videoId string) (string, error) {
	params := url.Values{}
	params.Add("id", videoId)
	apiUrl := "https://youtube-mp36.p.rapidapi.com/dl?" + params.Encode()

	req, _ := http.NewRequest(http.MethodGet, apiUrl, nil)
	req.Header.Add("X-RapidAPI-Host", "youtube-mp36.p.rapidapi.com")
	req.Header.Add("X-RapidAPI-Key", api.key)

	httpResponse, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer httpResponse.Body.Close()
	bodyData, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return "", err
	}
	var response model.ApiResponse
	json.Unmarshal(bodyData, &response)

	if response.Status != "ok" {
		return "", errors.New(response.Msg)
	}

	return response.Link, nil
}
