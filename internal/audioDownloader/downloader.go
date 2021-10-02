package audiodownloader

import (
	"errors"
	"io"
	"strings"

	"github.com/kkdai/youtube/v2"
)

func getAudioFormatIndex(video *youtube.Video) (int, error) {
	for i := range video.Formats {
		if strings.Contains(video.Formats[i].MimeType, "audio/mp4") {
			return i, nil
		}
	}

	return 0, errors.New("audio format was not found")
}

func DownloadAudio(videoId string) (io.Reader, error) {
	client := youtube.Client{}

	video, err := client.GetVideo(videoId)
	if err != nil {
		return nil, err
	}

	audioFormatIndex, err := getAudioFormatIndex(video)
	if err != nil {
		return nil, err
	}

	stream, _, err := client.GetStream(video, &video.Formats[audioFormatIndex])
	if err != nil {
		return nil, err
	}

	return stream, nil
}
