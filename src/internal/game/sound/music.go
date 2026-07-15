package sound

import (
	"bytes"
	"embed"
	"fmt"
	"io"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
)

//go:embed game_music.mp3 game_music_menu.mp3 beer_glass_hit.mp3 jump.mp3 laugh.mp3
var assets embed.FS

const (
	sampleRate  = 44100
	musicVolume = 0.45
	sfxVolume   = 0.8
)

type Manager struct {
	ctx            *audio.Context
	menuPlayer     *audio.Player
	gamePlayer     *audio.Player
	lifeLostPlayer *audio.Player
	jumpPlayer     *audio.Player
	gameOverPlayer *audio.Player
	menuMode       bool
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

	lifeLostPlayer, err := newOneShotPlayer(ctx, "beer_glass_hit.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		return nil, fmt.Errorf("life lost sfx: %w", err)
	}
	lifeLostPlayer.SetVolume(sfxVolume)

	jumpPlayer, err := newOneShotPlayer(ctx, "jump.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		lifeLostPlayer.Close()
		return nil, fmt.Errorf("jump sfx: %w", err)
	}
	jumpPlayer.SetVolume(sfxVolume)

	gameOverPlayer, err := newOneShotPlayer(ctx, "laugh.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		lifeLostPlayer.Close()
		jumpPlayer.Close()
		return nil, fmt.Errorf("game over sfx: %w", err)
	}
	gameOverPlayer.SetVolume(sfxVolume)

	return &Manager{
		ctx:            ctx,
		menuPlayer:     menuPlayer,
		gamePlayer:     gamePlayer,
		lifeLostPlayer: lifeLostPlayer,
		jumpPlayer:     jumpPlayer,
		gameOverPlayer: gameOverPlayer,
	}, nil
}

func decodePCM(name string) ([]byte, error) {
	data, err := assets.ReadFile(name)
	if err != nil {
		return nil, err
	}

	stream, err := mp3.DecodeWithoutResampling(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return io.ReadAll(stream)
}

func newLoopPlayer(ctx *audio.Context, name string) (*audio.Player, error) {
	pcm, err := decodePCM(name)
	if err != nil {
		return nil, err
	}

	const channels = 2
	bytesPerSample := 2 * channels
	blend := int64(sampleRate) * int64(bytesPerSample) / 10
	loopLen := int64(len(pcm)) - blend
	loop := audio.NewInfiniteLoop(bytes.NewReader(pcm), loopLen)
	return ctx.NewPlayer(loop)
}

func newOneShotPlayer(ctx *audio.Context, name string) (*audio.Player, error) {
	pcm, err := decodePCM(name)
	if err != nil {
		return nil, err
	}
	return ctx.NewPlayer(bytes.NewReader(pcm))
}

func (m *Manager) PlayJump() {
	if m == nil || m.jumpPlayer == nil {
		return
	}
	_ = m.jumpPlayer.Rewind()
	m.jumpPlayer.Play()
}

func (m *Manager) PlayLifeLost() {
	if m == nil || m.lifeLostPlayer == nil {
		return
	}
	_ = m.lifeLostPlayer.Rewind()
	m.lifeLostPlayer.Play()
}

func (m *Manager) PlayGameOver() {
	if m == nil || m.gameOverPlayer == nil {
		return
	}
	_ = m.gameOverPlayer.Rewind()
	m.gameOverPlayer.Play()
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
	if m.lifeLostPlayer != nil {
		if e := m.lifeLostPlayer.Close(); e != nil && err == nil {
			err = e
		}
	}
	if m.jumpPlayer != nil {
		if e := m.jumpPlayer.Close(); e != nil && err == nil {
			err = e
		}
	}
	if m.gameOverPlayer != nil {
		if e := m.gameOverPlayer.Close(); e != nil && err == nil {
			err = e
		}
	}
	return err
}
