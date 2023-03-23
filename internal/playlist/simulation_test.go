package playlist

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"sbercloud/internal/song"
	"testing"
	"time"
)

// Technical requirements

// TestBasicPlaylist_InteractionInterface
// Должен быть описан четко определенный интерфейс для взаимодействия с плейлистом
func TestBasicPlaylist_InteractionInterface(t *testing.T) {
	ctx := context.Background()
	playlist := NewSimulation(ctx)

	playlist.AddSong(song.NewBasic("A", time.Millisecond*100))
	playlist.AddSong(song.NewBasic("B", time.Millisecond*100))
	playlist.AddSong(song.NewBasic("C", time.Millisecond*100))
	require.EqualValues(t, 3, playlist.trackCount)

	playlist.Play()
	require.EqualValues(t, "A", playlist.current.song.Name())
	require.EqualValues(t, playing, playlist.state)

	playlist.AddSong(song.NewBasic("D", time.Millisecond*100))
	require.EqualValues(t, 4, playlist.trackCount)
	require.EqualValues(t, playing, playlist.state)

	playlist.Pause()
	require.EqualValues(t, "A", playlist.current.song.Name())
	require.EqualValues(t, paused, playlist.state)

	playlist.Play()
	require.EqualValues(t, "A", playlist.current.song.Name())
	require.EqualValues(t, playing, playlist.state)

	playlist.Next()
	require.EqualValues(t, "B", playlist.current.song.Name())
	require.EqualValues(t, playing, playlist.state)
	playlist.Next()
	require.EqualValues(t, "C", playlist.current.song.Name())
	require.EqualValues(t, playing, playlist.state)
	playlist.Next()
	require.EqualValues(t, "D", playlist.current.song.Name())
	require.EqualValues(t, playing, playlist.state)
	playlist.Next()
	require.EqualValues(t, idle, playlist.state)

	playlist.Prev()
	require.EqualValues(t, "D", playlist.current.song.Name())
	require.EqualValues(t, playing, playlist.state)
	playlist.Prev()
	require.EqualValues(t, "C", playlist.current.song.Name())
	require.EqualValues(t, playing, playlist.state)
	playlist.Prev()
	require.EqualValues(t, "B", playlist.current.song.Name())
	require.EqualValues(t, playing, playlist.state)
	playlist.Prev()
	require.EqualValues(t, "A", playlist.current.song.Name())
	require.EqualValues(t, playing, playlist.state)
	playlist.Prev()
	require.EqualValues(t, idle, playlist.state)
}

// TestBasicPlaylist_DoubleLinkedList
// Плейлист должен быть реализован с использованием двусвязного списка
func TestBasicPlaylist_DoubleLinkedList(t *testing.T) {
	t.Run("sequential access forward", func(t *testing.T) {
		pl := threeSongsPlaylist()
		assert.Equal(t, "alpha", pl.current.song.Name())
		pl.Next()
		assert.Equal(t, "beta", pl.current.song.Name())
		pl.Next()
		assert.Equal(t, "charlie", pl.current.song.Name())
	})
	t.Run("no forward cycle", func(t *testing.T) {
		pl := threeSongsPlaylist()
		pl.Next()
		pl.Next()
		pl.Next()
		assert.Zero(t, pl.current)
	})
	t.Run("sequential access backward", func(t *testing.T) {
		pl := threeSongsPlaylist()
		pl.Next()
		pl.Next()
		assert.Equal(t, "charlie", pl.current.song.Name())
		pl.Prev()
		assert.Equal(t, "beta", pl.current.song.Name())
		pl.Prev()
		assert.Equal(t, "alpha", pl.current.song.Name())
	})
	t.Run("no backward cycle", func(t *testing.T) {
		pl := threeSongsPlaylist()
		pl.Prev()
		assert.Zero(t, pl.current)
	})

}

// TestBasicPlaylist_AllSongsHaveDurationProperty
// Каждая песня в плейлисте должна иметь свойство Duration
func TestBasicPlaylist_AllSongsHaveDurationProperty(t *testing.T) {
	t.Run("song interface has duration", func(t *testing.T) {
		addSongMethod, addSongExists := reflect.TypeOf((*Playlist)(nil)).Elem().MethodByName("AddSong")
		require.True(t, addSongExists)

		_, isDurationMethodExists := addSongMethod.Type.In(0).MethodByName("Duration")
		assert.True(t, isDurationMethodExists)
	})
}

// TestBasicPlaylist_PlayingDoesNotBlockInteraction
// Воспроизведение песни не должно блокировать методы управления.
func TestBasicPlaylist_PlayingDoesNotBlockInteraction(t *testing.T) {
	t.Run("control operations work during playing", func(t *testing.T) {
		pl := threeSongsPlaylist()
		pl.Play()
		assert.Equal(t, "alpha", pl.current.song.Name())
		assert.Equal(t, playing, pl.state)

		pl.Next()
		assert.Equal(t, "beta", pl.current.song.Name())
		assert.Equal(t, playing, pl.state)

		pl.Prev()
		assert.Equal(t, "alpha", pl.current.song.Name())
		assert.Equal(t, playing, pl.state)

		pl.AddSong(song.NewBasic("any", time.Hour))
		assert.Equal(t, 4, pl.trackCount)
	})
}

// TestBasicPlaylist_PayingDurationLimitedBySong
// Метод воспроизведения должен начать воспроизведение с длительностью,
// ограниченной свойством Duration песни.
// Воспроизведение должно эмулироваться длительной операцией.
func TestBasicPlaylist_PayingDurationLimitedBySong(t *testing.T) {
	t.Run("playing is limited by song duration", func(t *testing.T) {
		songDuration := time.Millisecond * 100
		single := song.NewBasic("A", songDuration)
		pl := NewSimulation(context.Background())
		pl.AddSong(single)

		pl.Play()

		assert.Equal(t, playing, pl.state)
		<-time.Tick(songDuration + time.Millisecond)
		assert.Equal(t, idle, pl.state)
	})
	t.Run("total playing is limited by song duration", func(t *testing.T) {
		songDuration := time.Millisecond * 100
		songs := []Song{
			song.NewBasic("A", songDuration),
			song.NewBasic("B", songDuration),
			song.NewBasic("C", songDuration),
		}

		pl := NewSimulation(context.Background())
		for _, s := range songs {
			pl.AddSong(s)
		}
		pl.Play()

		assert.Equal(t, playing, pl.state)
		switchSongDuration := time.Millisecond
		waitFor := time.Duration(len(songs)) * (songDuration + switchSongDuration)
		<-time.Tick(waitFor)
		assert.Equal(t, idle, pl.state)
	})
}

// TestBasicPlaylist_TracksSwitchingAutomatically
// Следующая песня должна воспроизводиться автоматически после окончания текущей песни
func TestBasicPlaylist_TracksSwitchingAutomatically(t *testing.T) {
	t.Run("playlist play songs successively", func(t *testing.T) {
		songDuration := time.Millisecond * 100
		songs := []Song{
			song.NewBasic("A", songDuration),
			song.NewBasic("B", songDuration),
			song.NewBasic("C", songDuration),
		}

		pl := NewSimulation(context.Background())
		for _, s := range songs {
			pl.AddSong(s)
		}
		pl.Play()

		assert.Equal(t, playing, pl.state)
		for _, s := range songs {
			assert.Equal(t, s.Name(), pl.current.song.Name())
			someTimeForSwitch := time.Millisecond
			<-time.Tick(songDuration + someTimeForSwitch)
		}
	})
}

// TestBasicPlaylist_ResumeAfterPause
// Метод Pause должен приостановить текущее воспроизведение,
// и когда воспроизведение вызывается снова, оно должно продолжаться с момента паузы.
func TestBasicPlaylist_ResumeAfterPause(t *testing.T) {
	t.Run("pause", func(t *testing.T) {
		pl := NewSimulation(context.Background())
		songDuration := time.Second
		pl.AddSong(song.NewBasic("Song", songDuration))
		pl.Play()
		assert.Equal(t, playing, pl.state)
		playTime := time.Millisecond * 70
		<-time.Tick(playTime)
		pl.Pause()
		assert.Equal(t, songDuration-playTime, pl.left)
	})
	t.Run("pause and resume", func(t *testing.T) {
		pl := NewSimulation(context.Background())
		songDuration := time.Millisecond * 100
		pl.AddSong(song.NewBasic("Song", songDuration))
		pl.Play()
		assert.Equal(t, playing, pl.state)
		playTime := time.Millisecond * 70
		<-time.Tick(playTime)
		pl.Pause()
		left := songDuration - playTime
		assert.Equal(t, paused, pl.state)
		assert.GreaterOrEqual(t, left, pl.left)
		pl.Play()
		assert.Equal(t, playing, pl.state)
		playTime2 := time.Millisecond * 20
		<-time.Tick(playTime2)
		pl.Pause()
		assert.Equal(t, paused, pl.state)
		//assert.GreaterOrEqual(t, songDuration-playTime2-playTime, pl.left) // todo fix
	})
}

// TestBasicPlaylist_AddSongAddsToEnd
// Метод AddSong должен добавить новую песню в конец списка.
func TestBasicPlaylist_AddSongAddsToEnd(t *testing.T) {
	t.Run("added song is last", func(t *testing.T) {
		pl := NewSimulation(context.Background())
		pl.AddSong(song.NewBasic("A", time.Millisecond))
		assert.Equal(t, "A", pl.last.song.Name())
		pl.AddSong(song.NewBasic("B", time.Millisecond))
		assert.Equal(t, "B", pl.last.song.Name())
	})
}

// TestBasicPlaylist_Next
// Вызов метода Next должен начать воспроизведение следущей песни.
// Таким образом текущее воспроизведение должно быть остановлено
// и начато воспроизведение следущей песни.
func TestBasicPlaylist_Next(t *testing.T) {
	// todo observer pattern for statuses
	t.Run("next is stopping switching resuming", func(t *testing.T) {
		pl := NewSimulation(context.Background())
		pl.AddSong(song.NewBasic("A", time.Millisecond))
		pl.AddSong(song.NewBasic("B", time.Millisecond))
		pl.Play()
		assert.Equal(t, playing, pl.state)
		pl.Next()
		assert.Equal(t, playing, pl.state)
	})
}

// TestBasicPlaylist_Prev
// Вызов метода Prev должен остановить текущее воспроизведение
// и начать воспроизведение предыдущей песни.
func TestBasicPlaylist_Prev(t *testing.T) {
	// todo observer pattern for statuses
	t.Run("prev is stopping switching resuming", func(t *testing.T) {
		pl := NewSimulation(context.Background())
		pl.AddSong(song.NewBasic("A", time.Millisecond))
		pl.AddSong(song.NewBasic("B", time.Millisecond))
		pl.AddSong(song.NewBasic("C", time.Millisecond))
		pl.Play()
		assert.Equal(t, playing, pl.state)
		assert.Equal(t, "A", pl.current.song.Name())
		pl.Next()
		assert.Equal(t, playing, pl.state)
		assert.Equal(t, "B", pl.current.song.Name())
		pl.Prev()
		assert.Equal(t, playing, pl.state)
		assert.Equal(t, "A", pl.current.song.Name())
	})
}

// TestBasicPlaylist_AddSongIsConcurrent
// Реализация метода AddSong должна проводиться с учетом
// одновременного, конкурентного доступа.
func TestBasicPlaylist_AddSongIsConcurrent(t *testing.T) {
	// todo ask what tests to add
	pl := NewSimulation(context.Background())
	client := func(playlist Playlist, songsCount int, clientName string, done chan struct{}) {
		for i := 0; i < songsCount; i++ {
			playlist.AddSong(song.NewBasic(fmt.Sprintf("%s_%d", clientName, i), time.Nanosecond))
		}
		done <- struct{}{}
	}

	clientsCount := 2
	doneC := make(chan struct{}, clientsCount)
	songsPerClient := 10000
	for i := 0; i < clientsCount; i++ {
		go client(pl, songsPerClient, fmt.Sprintf("%d", i), doneC)
	}
	for i := 0; i < clientsCount; i++ {
		<-doneC
	}

	assert.Equal(t, 2*songsPerClient, pl.trackCount)
}

// TestBasicPlaylist_PauseIsConcurrentSafe
// Следует учитывать, что воспроизведение может быть остановлено извне.
func TestBasicPlaylist_PauseIsConcurrentSafe(t *testing.T) {
	// todo ask how to do
}

// TestBasicPlaylist_NextAndPrevAreConcurrentSafe
// Реализация методов Next/Prev должна проводиться с учетом одновременного, конкурентного доступа.
func TestBasicPlaylist_NextAndPrevAreConcurrentSafe(t *testing.T) {
	// todo ask how to do
}

func threeSongsPlaylist() *Simulation {
	pl := NewSimulation(context.Background())

	pl.AddSong(song.NewBasic("alpha", time.Nanosecond))
	pl.AddSong(song.NewBasic("beta", time.Nanosecond))
	pl.AddSong(song.NewBasic("charlie", time.Nanosecond))
	return pl
}
