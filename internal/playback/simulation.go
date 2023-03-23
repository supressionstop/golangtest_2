package playback

import (
	"log"
	"time"
)

const defaultRate = 10 * time.Millisecond

func SimulatePlayback(changer chan time.Duration, interruptC chan struct{}) (stoppedAt <-chan time.Duration, finished <-chan struct{}) {
	stoppedAtNotifier := make(chan time.Duration)
	finishedNotifier := make(chan struct{})

	go playback(changer, interruptC, stoppedAtNotifier, finishedNotifier)

	return stoppedAtNotifier, finishedNotifier
}

func playback(changerC chan time.Duration, interruptC chan struct{}, stoppedAt chan time.Duration, finished chan struct{}) {
	ticker := time.NewTicker(defaultRate)
	var left time.Duration

	for {
		select {
		case d := <-changerC:
			log.Println("playback change", d)
			left = d
			ticker.Reset(defaultRate)
		case <-interruptC:
			log.Print("playback interrupt")
			ticker.Stop()
			stoppedAt <- left
		case <-ticker.C:
			left -= defaultRate
			log.Print(".")
			if left <= 0 {
				ticker.Stop()
				finished <- struct{}{}
			}
		}
	}
}
