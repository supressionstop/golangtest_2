package main

import (
	"context"
	"sbercloud/internal/playlist"
)

func main() {
	waitC := make(chan struct{})
	ctx := context.Background()

	pl := playlist.NewSimulation(ctx)
	pl.Play()

	<-waitC
}
