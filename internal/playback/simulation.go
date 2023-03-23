package playback

import (
	"log"
	"time"
)

const defaultRate = 10 * time.Millisecond

func SimPbV2(changer chan time.Duration, interruptC chan struct{}) (stoppedAt <-chan time.Duration, finished <-chan struct{}) {
	stoppedAtNotifier := make(chan time.Duration)
	finishedNotifier := make(chan struct{})

	go pb(changer, interruptC, stoppedAtNotifier, finishedNotifier)

	return stoppedAtNotifier, finishedNotifier
}

func pb(changerC chan time.Duration, interruptC chan struct{}, stoppedAt chan time.Duration, finished chan struct{}) {
	ticker := time.NewTicker(defaultRate)
	var left time.Duration

	for {
		select {
		case d := <-changerC:
			log.Println("pb change", d)
			left = d
			ticker.Reset(defaultRate)
		case <-interruptC:
			log.Print("pb interrupt")
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

func SimulatePlayback(d time.Duration, interruptC <-chan struct{}) (stoppedAt <-chan time.Duration, finished chan struct{}) {
	stoppedAtNotifier := make(chan time.Duration)
	finishedNotifier := make(chan struct{})
	rate := defaultRate

	if d <= rate {
		rate = d
	}

	go playback(rate, d, interruptC, stoppedAtNotifier, finishedNotifier)

	return stoppedAtNotifier, finishedNotifier
}

func playback(rate, d time.Duration, interruptC <-chan struct{}, stoppedAt chan<- time.Duration, finished chan<- struct{}) {
	ticker := time.NewTicker(rate)
	left := d

	for left > 0 {
		select {
		case <-interruptC:
			log.Println("interrupt", left)
			stoppedAt <- left
			return
		default:
			<-ticker.C
			left -= rate
			log.Print(".")
		}
	}
	// finished
	finished <- struct{}{}
}
