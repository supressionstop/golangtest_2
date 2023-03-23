package song

import "time"

type Basic struct {
	name     string
	duration time.Duration
}

func NewBasic(name string, duration time.Duration) Basic {
	return Basic{
		name:     name,
		duration: duration,
	}
}

func (s Basic) Name() string {
	return s.name
}

func (s Basic) Duration() time.Duration {
	return s.duration
}
