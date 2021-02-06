package youtubebot

type RestResponse struct {
	Items []Item `json:"items"`
}

type Item struct {
	Id ItemInfo `json:"id"`
}

type ItemInfo struct {
	VideoId string `json:"videoId"`
}
