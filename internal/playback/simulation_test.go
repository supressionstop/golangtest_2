package playback

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSimulatePlayback_WholeDuration(t *testing.T) {
	testCasesWholeDuration := []time.Duration{
		time.Millisecond,
		time.Millisecond * 2,
		time.Millisecond * 3,
		time.Millisecond * 5,
		time.Millisecond * 7,
		time.Millisecond * 9,
		time.Millisecond * 21,
		time.Millisecond * 77,
		time.Millisecond * 501,
		time.Millisecond * 999,
		time.Second,
	}
	for _, testDuration := range testCasesWholeDuration {
		testWholeDuration(t, testDuration)
	}
}

func testWholeDuration(t *testing.T, testDuration time.Duration) {
	t.Helper()

	t.Run(fmt.Sprintf("%q whole duration played precisely", testDuration), func(t *testing.T) {
		t.Parallel()
		// arrange
		stopC := make(chan struct{})
		duration := testDuration
		epsilon := defaultRate // tolerated error

		// act
		_, finished := SimulatePlayback(duration, stopC)
		startedAt := time.Now()
		left, isOpen := <-finished
		finishedAt := time.Now()

		// assert
		require.Zero(t, left)
		require.True(t, isOpen)
		howLongDidItTake := finishedAt.Sub(startedAt) - duration
		require.LessOrEqual(t, howLongDidItTake, epsilon)
	})
}

func TestSimulatePlayback_Interruption(t *testing.T) {
	testCasesInterruption := []struct {
		duration  time.Duration
		stopAfter time.Duration
	}{
		{time.Millisecond * 200, time.Millisecond * 40},
		{time.Millisecond * 240, time.Millisecond * 10},
	}
	for _, tc := range testCasesInterruption {
		tca := tc
		testInterruption(t, tca.duration, tca.stopAfter)
	}
}

func testInterruption(t *testing.T, testDuration, stopAfter time.Duration) {
	t.Helper()

	t.Run(fmt.Sprintf("duration %s interrupted after %s works", testDuration, stopAfter), func(t *testing.T) {
		t.Parallel()
		if stopAfter > testDuration {
			require.GreaterOrEqual(t, testDuration, stopAfter)
		}

		// arrange
		stopC := make(chan struct{}, 1)
		duration := testDuration
		epsilon := defaultRate // tolerated error

		// act
		stoppedAt, _ := SimulatePlayback(duration, stopC)
		startedAt := time.Now()
		<-time.Tick(stopAfter)
		stopC <- struct{}{}
		left, isOpen := <-stoppedAt
		finishedAt := time.Now()

		// assert
		require.EqualValues(t, duration-stopAfter-defaultRate, left)
		require.True(t, isOpen)
		howLongDidItTake := finishedAt.Sub(startedAt) - duration
		require.LessOrEqual(t, howLongDidItTake, epsilon)
	})
}

func TestSimulatePlayback_WholeDurationV2(t *testing.T) {
	testCasesWholeDuration := []time.Duration{
		time.Millisecond,
		time.Millisecond * 2,
		time.Millisecond * 3,
		time.Millisecond * 5,
		time.Millisecond * 7,
		time.Millisecond * 9,
		time.Millisecond * 21,
		time.Millisecond * 77,
		time.Millisecond * 501,
		time.Millisecond * 999,
		time.Second,
	}
	for _, testDuration := range testCasesWholeDuration {
		testWholeDurationV2(t, testDuration)
	}
}

func testWholeDurationV2(t *testing.T, testDuration time.Duration) {
	t.Helper()

	t.Run(fmt.Sprintf("%q whole duration played precisely", testDuration), func(t *testing.T) {
		t.Parallel()
		// arrange
		changerC := make(chan time.Duration)
		stopC := make(chan struct{})
		epsilon := defaultRate // tolerated error

		// act
		_, finished := SimPbV2(changerC, stopC)
		changerC <- testDuration
		startedAt := time.Now()
		left, isOpen := <-finished
		finishedAt := time.Now()

		// assert
		require.Zero(t, left)
		require.True(t, isOpen)
		howLongDidItTake := finishedAt.Sub(startedAt) - testDuration
		require.LessOrEqual(t, howLongDidItTake, epsilon)
	})
}

func TestSimulatePlayback_InterruptionV2(t *testing.T) {
	testCasesInterruption := []struct {
		duration  time.Duration
		stopAfter time.Duration
	}{
		{time.Millisecond * 200, time.Millisecond * 40},
		{time.Millisecond * 240, time.Millisecond * 10},
	}
	for _, tc := range testCasesInterruption {
		tca := tc
		testInterruptionV2(t, tca.duration, tca.stopAfter)
	}
}

func testInterruptionV2(t *testing.T, testDuration, stopAfter time.Duration) {
	t.Helper()

	t.Run(fmt.Sprintf("duration %s interrupted after %s works", testDuration, stopAfter), func(t *testing.T) {
		t.Parallel()
		if stopAfter > testDuration {
			require.GreaterOrEqual(t, testDuration, stopAfter)
		}

		// arrange
		changerC := make(chan time.Duration)
		stopC := make(chan struct{}, 1)
		epsilon := defaultRate // tolerated error

		// act
		stoppedAt, _ := SimPbV2(changerC, stopC)
		changerC <- testDuration
		startedAt := time.Now()
		<-time.Tick(stopAfter)
		stopC <- struct{}{}
		left, isOpen := <-stoppedAt
		finishedAt := time.Now()

		// assert
		require.EqualValues(t, testDuration-stopAfter, left)
		require.True(t, isOpen)
		howLongDidItTake := finishedAt.Sub(startedAt) - testDuration
		require.LessOrEqual(t, howLongDidItTake, epsilon)
	})
}
