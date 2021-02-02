package YouTubeBot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const YoutubeSearchUrl = "https://www.googleapis.com/youtube/v3/search"
const YoutubeApiToken = "AIzaSyDhqepLthlr88MTnhOWd8sXoeT5uL5Ki_c"
const YoutubeVideoUrl = "https://www.youtube.com/watch?v="

//GET https://youtube.googleapis.com/youtube/v3/search?part=id&channelId=UCDqlBCW2D1CD8m49YDoEv4g&maxResults=1&order=date&key=[YOUR_API_KEY] HTTP/1.1

func GetLastVideo(channelUrl string) (string, error) {
	items, err := retrieveVideo(channelUrl)
	if err != nil {
		return "", err
	}

	if len(items) < 1 {
		return "", errors.New("No video found")
	}

	return YoutubeVideoUrl + items[0].Id.VideoId, nil
}

func retrieveVideo(channelUrl string) ([]Item, error) {
	req, err := makeRequest(channelUrl, 1)
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var restResponse RestResponse
	err = json.Unmarshal(body, &restResponse)
	if err != nil {
		return nil, err
	}

	return restResponse.Items, nil
}

func makeRequest(channelUrl string, maxResults int) (*http.Request, error) {
	lastSlashIndex := strings.LastIndex(channelUrl, "/")
	channelId := channelUrl[lastSlashIndex + 1 :]

	req, err := http.NewRequest("GET", YoutubeSearchUrl, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("part", "id")
	query.Add("channelId", channelId)
	query.Add("maxResults", strconv.Itoa(maxResults))
	query.Add("order", "date")
	query.Add("key", YoutubeApiToken)

	req.URL.RawQuery = query.Encode()
	fmt.Print(req.URL.String())
	return req, nil
}