package model

type YoutubeItem struct {
	Title        string
	ThumbnailURL string
	Id           string
}

type YoutubePlaylist struct {
	YoutubeItem
	ItemCount uint
}
