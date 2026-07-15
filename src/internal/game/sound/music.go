package sound

import (
	"bytes"
	"embed"
	"fmt"
	"io"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
)

//go:embed game_music.mp3 game_music_menu.mp3
var assets embed.FS

const (
	sampleRate  = 44100
	musicVolume = 0.45
)

type Manager struct {
	ctx        *audio.Context
	menuPlayer *audio.Player
	gamePlayer *audio.Player
	menuMode   bool
}

func NewManager() (*Manager, error) {
	ctx := audio.NewContext(sampleRate)

	menuPlayer, err := newLoopPlayer(ctx, "game_music_menu.mp3")
	if err != nil {
		return nil, fmt.Errorf("menu music: %w", err)
	}
	menuPlayer.SetVolume(musicVolume)

	gamePlayer, err := newLoopPlayer(ctx, "game_music.mp3")
	if err != nil {
		menuPlayer.Close()
		return nil, fmt.Errorf("game music: %w", err)
	}
	gamePlayer.SetVolume(musicVolume)

	return &Manager{
		ctx:        ctx,
		menuPlayer: menuPlayer,
		gamePlayer: gamePlayer,
	}, nil
}

func newLoopPlayer(ctx *audio.Context, name string) (*audio.Player, error) {
	data, err := assets.ReadFile(name)
	if err != nil {
		return nil, err
	}

	stream, err := mp3.DecodeWithoutResampling(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	pcm, err := io.ReadAll(stream)
	if err != nil {
		return nil, err
	}

	const channels = 2
	bytesPerSample := 2 * channels
	blend := int64(sampleRate) * int64(bytesPerSample) / 10
	loopLen := int64(len(pcm)) - blend
	loop := audio.NewInfiniteLoop(bytes.NewReader(pcm), loopLen)
	player, err := ctx.NewPlayer(loop)
	if err != nil {
		return nil, err
	}
	return player, nil
}

func (m *Manager) SetMode(menu bool) {
	if m == nil || m.menuMode == menu {
		return
	}
	m.menuMode = menu

	if menu {
		m.gamePlayer.Pause()
		_ = m.menuPlayer.Rewind()
		m.menuPlayer.Play()
		return
	}

	m.menuPlayer.Pause()
	_ = m.gamePlayer.Rewind()
	m.gamePlayer.Play()
}

func (m *Manager) Close() error {
	if m == nil {
		return nil
	}
	var err error
	if m.menuPlayer != nil {
		if e := m.menuPlayer.Close(); e != nil && err == nil {
			err = e
		}
	}
	if m.gamePlayer != nil {
		if e := m.gamePlayer.Close(); e != nil && err == nil {
			err = e
		}
	}
	return err
}
