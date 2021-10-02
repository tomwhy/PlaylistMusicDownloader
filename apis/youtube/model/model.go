package model

type YoutubeItem struct {
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnail_url"`
	Id           string `json:"id"`
}

type YoutubePlaylist struct {
	YoutubeItem
	ItemCount uint
}

type YoutubeVideo struct {
	YoutubeItem
	DownloadUrl string `json:"download_url"`
}
