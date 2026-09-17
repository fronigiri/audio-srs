package audio

import (
	"os"
	"path/filepath"

	"github.com/dhowden/tag"
)

type Song struct {
	Title  string
	Artist string
	Album  string
	Path   string
}

func ScanLibrary(libraryPath string) (map[string][]Song, error) {
	albums := make(map[string][]Song)

	err := filepath.Walk(libraryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		ext := filepath.Ext(path)
		if ext == ".mp3" || ext == ".flac" || ext == ".m4a" {
			file, err := os.Open(path)
			if err != nil {
				return nil // Skip files we cannot open
			}
			defer file.Close()

			m, err := tag.ReadFrom(file)
			if err != nil {
				return nil // Skip files with unreadable tags
			}

			albumName := m.Album()
			if albumName == "" {
				albumName = "Unknown Album"
			}

			song := Song{
				Title:  m.Title(),
				Artist: m.Artist(),
				Album:  albumName,
				Path:   path,
			}

			albums[albumName] = append(albums[albumName], song)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return albums, nil
}
