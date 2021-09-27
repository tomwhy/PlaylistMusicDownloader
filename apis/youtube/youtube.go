package youtube

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/tomwhy/PlaylistMusicDownloader/apis/youtube/model"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const maxPageSize = 50

type YoutubeAPI struct {
	ctx     context.Context
	service *youtube.Service

	logger *logrus.Logger
}

func NewYoutubeAPI(options ...option.ClientOption) *YoutubeAPI {
	api := new(YoutubeAPI)
	api.ctx = context.Background()
	api.logger = logrus.New()

	var err error
	api.service, err = youtube.NewService(api.ctx, options...)

	if err != nil {
		api.logger.Error("Failed to create youtube service", err)
		return nil
	}

	return api
}

func (api *YoutubeAPI) GetAllPlaylists(page string, pageSize uint) (playlists []model.YoutubePlaylist, nextPage string, prevPage string, err error) {
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	request := api.service.Playlists.List([]string{"id", "snippet", "contentDetails"}).Mine(true).MaxResults(int64(pageSize))
	if page != "" {
		request = request.PageToken(page)
	}

	res, err := request.Do()
	if err != nil {
		logrus.Error("Failed getting playlists.", err)
		return nil, "", "", err
	}

	for _, item := range res.Items {
		playlists = append(playlists, model.YoutubePlaylist{
			YoutubeItem: model.YoutubeItem{
				Title:        item.Snippet.Title,
				ThumbnailURL: item.Snippet.Thumbnails.Default.Url,
				Id:           item.Id,
			},
			ItemCount: uint(item.ContentDetails.ItemCount),
		})
	}

	return playlists, res.NextPageToken, res.PrevPageToken, nil
}

func (api *YoutubeAPI) GetPlaylistSongs(playlistId, page string) (videos []model.YoutubeVideo, nextPage string, err error) {
	request := api.service.PlaylistItems.List([]string{"contentDetails", "snippet"}).PlaylistId(playlistId).MaxResults(50)
	if page != "" {
		request.PageToken(page)
	}

	response, err := request.Do()
	if err != nil {
		return nil, "", err
	}

	for _, item := range response.Items {
		videos = append(videos, model.YoutubeVideo{
			YoutubeItem: model.YoutubeItem{
				Title:        item.Snippet.Title,
				ThumbnailURL: item.Snippet.Thumbnails.Default.Url,
				Id:           item.ContentDetails.VideoId,
			},
		})
	}

	return videos, response.NextPageToken, nil
}
