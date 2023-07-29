package node

import (
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
	"github.com/kmtym1998/es-indexer/elasticsearch"
	"github.com/samber/lo"
)

type AudioFile struct {
	FilePath        string   `json:"filePath"`
	FileName        string   `json:"fileName"`
	Artists         []string `json:"artists"`
	AlbumArtist     string   `json:"albumArtist"`
	Album           string   `json:"album"`
	Title           string   `json:"title"`
	Tags            []string `json:"tags"`
	ContainedTracks []string `json:"containedTracks"`
}

type AudioFileList []AudioFile

var ErrNotAudioFile = errors.New("not audio file")

func isM4A(mimeType string) bool {
	for _, possibleMimeType := range []string{
		"audio/aac",
		"audio/aacp",
		"audio/3gpp",
		"audio/3gpp2",
		"audio/mp4",
		"audio/MP4A-LATM",
		"audio/mpeg4-generic",
		"video/mp4",
	} {
		if mimeType == possibleMimeType {
			return true
		}
	}

	return false
}

func splitArtists(str string) (results []string) {
	// これリファクタしたい
	for _, s1 := range strings.Split(str, ",") {
		for _, s2 := range strings.Split(s1, "&") {
			for _, s3 := range strings.Split(s2, " x ") {
				for _, s4 := range strings.Split(s3, "X") {
					for _, s5 := range strings.Split(s4, " vs ") {
						for _, s6 := range strings.Split(s5, " feat. ") {
							results = append(results, strings.TrimSpace(s6))
						}
					}
				}
			}
		}
	}

	return results
}

func NewAudioFileNode(audioFilePath string) (*AudioFile, error) {
	f, err := os.Open(audioFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tag, err := tag.ReadFrom(f)
	if err == nil {
		return &AudioFile{
			FilePath:        audioFilePath,
			FileName:        filepath.Base(audioFilePath),
			Artists:         splitArtists(tag.Artist()),
			AlbumArtist:     tag.AlbumArtist(),
			Album:           tag.Album(),
			Title:           tag.Title(),
			Tags:            strings.Split(tag.Genre(), ","),
			ContainedTracks: []string{tag.Title()},
		}, nil
	}

	b, err := os.ReadFile(audioFilePath)
	if err != nil {
		return nil, err
	}

	mimeType := http.DetectContentType(b)

	if isM4A(mimeType) || mimeType == "audio/wave" || filepath.Ext(audioFilePath) == ".wav" || filepath.Ext(audioFilePath) == ".mp3" {
		return &AudioFile{
			FilePath:        audioFilePath,
			FileName:        filepath.Base(audioFilePath),
			Artists:         []string{},
			Tags:            []string{},
			ContainedTracks: []string{},
		}, nil
	}

	return nil, ErrNotAudioFile
}

func (a AudioFile) NodeIdentifier() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(a.FilePath)))
}

func (a AudioFileList) IndexName() string {
	return "audio_files"
}

func (a AudioFileList) ToList() []elasticsearch.DocumentNode {
	return lo.Map(a, func(item AudioFile, _ int) elasticsearch.DocumentNode {
		return &item
	})
}
