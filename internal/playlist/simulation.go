package playlist

import (
	"context"
	"log"
	"sbercloud/internal/playback"
	"sync"
	"time"
)

type Simulation struct {
	ctx context.Context

	first      *track
	last       *track
	current    *track
	trackCount int

	state state
	left  time.Duration
	track track

	changerC   chan time.Duration
	interruptC chan struct{}
	finished   <-chan struct{}
	stoppedAt  <-chan time.Duration

	mu sync.Mutex
}

func NewSimulation(ctx context.Context) *Simulation {
	pl := &Simulation{
		ctx:        ctx,
		state:      idle,
		changerC:   make(chan time.Duration),
		interruptC: make(chan struct{}),
	}

	pl.stoppedAt, pl.finished = playback.SimPbV2(pl.changerC, pl.interruptC)
	go pl.watchForFinish()

	return pl
}

func (p *Simulation) AddSong(song Song) {
	p.mu.Lock()
	defer p.mu.Unlock()

	t := newTrack(song)
	if p.first == nil {
		p.first = t
		p.last = t
		p.current = p.first
		p.left = p.current.duration()
	} else {
		t.prev = p.last
		p.last.next = t
		p.last = t
	}
	p.trackCount++
}

func (p *Simulation) Play() {
	if p.isPlaying() {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.current == nil {
		p.current = p.first
		p.left = p.current.duration()
	}
	//p.stoppedAt, p.finished = playback.SimulatePlayback(p.current.duration(), p.interruptC)
	p.changerC <- p.left
	p.state = playing
	log.Println("play", p.current.song.Name(), p.current.song.Duration())
}

func (p *Simulation) Pause() {
	if !p.isPlaying() {
		return
	}
	log.Println("got pause")
	p.interruptC <- struct{}{}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = paused
	p.left = <-p.stoppedAt
	log.Println("left", p.left)
}

func (p *Simulation) Next() {
	if p.current == nil {
		p.Play()
		return
	}
	if p.current.next == nil {
		p.mu.Lock()
		p.state = idle
		p.current = nil
		p.mu.Unlock()
		log.Println("no next songs")
		return
	}
	p.Pause()
	p.nextSong()
	p.Play()
}

func (p *Simulation) Prev() {
	if p.current == nil {
		p.mu.Lock()
		p.current = p.last
		p.mu.Unlock()
		p.Play()
		return
	}
	if p.current.prev == nil {
		p.mu.Lock()
		p.state = idle
		p.current = nil
		p.mu.Unlock()
		log.Println("no prev songs")
		return
	}
	p.Pause()
	p.prevSong()
	p.Play()
}

func (p *Simulation) isPlaying() bool {
	return p.state == playing
}

func (p *Simulation) nextSong() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = switchingNext
	p.current = p.current.next
	log.Println("next")
}

func (p *Simulation) prevSong() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = switchingPrev
	p.current = p.current.prev
	log.Println("prev")
}

func (p *Simulation) watchForFinish() {
	// todo context done
	for range p.finished {
		log.Println("watch got finish")
		if p.current == nil || p.current.next == nil {
			log.Println("no next songs")
			p.mu.Lock()
			p.state = idle
			p.current = nil
			p.mu.Unlock()
			continue
		}
		p.nextSong()
		p.Play()
	}
}

type track struct {
	song Song
	prev *track
	next *track
}

func newTrack(song Song) *track {
	return &track{song: song}
}

func (t track) duration() time.Duration {
	return t.song.Duration()
}

type state string

var (
	idle          = state("idle")
	playing       = state("playing")
	paused        = state("paused")
	switchingNext = state("switchingNext")
	switchingPrev = state("switchingPrev")
)
