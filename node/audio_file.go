package node

import (
	"crypto/md5"
	"fmt"

	"github.com/kmtym1998/es-indexer/elasticsearch"
	"github.com/samber/lo"
)

type AudioFile struct {
	FilePath        string   `json:"filePath"`
	FileName        string   `json:"fileName"`
	Artists         []string `json:"artists"`
	Album           string   `json:"album"`
	Title           string   `json:"title"`
	Tags            []string `json:"tags"`
	ContainedTracks []string `json:"containedTracks"`
}

type AudioFileList []AudioFile

func (a AudioFile) NodeIdentifier() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(a.FilePath)))
}

func (a AudioFileList) IndexName() string {
	return "audio_files"
}

func (a AudioFileList) ToList() []elasticsearch.DocumentNode {
	return lo.Map(a, func(item AudioFile, _ int) elasticsearch.DocumentNode {
		return item
	})
}
