package playlist

import "time"

type Playlist interface {
	Play()
	Pause()
	AddSong(Song)
	Next()
	Prev()
}

type Song interface {
	Name() string
	Duration() time.Duration
}
