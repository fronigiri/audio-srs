package audio

import (
	"fmt"
	"log"
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

func SongBrowser(library string) error {

	libraryPath := library

	// Map to group songs by Album name: map[AlbumName][]Song
	albums := make(map[string][]Song)

	err := filepath.Walk(libraryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Process only audio files (e.g., .mp3, .flac, .m4a)
		ext := filepath.Ext(path)
		if ext == ".mp3" || ext == ".flac" || ext == ".m4a" {

			// 1. Open the file handle
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			// 2. Read metadata tags from the open file
			m, err := tag.ReadFrom(file)
			if err != nil {
				// Failed to read tags; skip or handle gracefully
				return err
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

			// 3. Add to the album bucket
			albums[albumName] = append(albums[albumName], song)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error scanning library: %v", err)
	}

	// Print out the grouped results
	for album, songs := range albums {
		fmt.Printf("Album: %s (%d tracks)\n", album, len(songs))
		for _, s := range songs {
			fmt.Printf("  - %s by %s\n", s.Title, s.Artist)
		}
	}
	return nil
}
