package sound

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
)

//go:embed game_music.mp3 game_music_menu.mp3 game_music_boss_1.mp3 beer_glass_hit.mp3 jump.mp3 laugh.mp3 hey1.mp3 hey2.mp3 power_up.mp3
var assets embed.FS

const (
	sampleRate     = 44100
	musicVolume    = 0.45
	sfxVolume      = 0.8
	passDadHeyProb = 0.33
)

type MusicMode int

const (
	MusicMenu MusicMode = iota
	MusicGame
	MusicBoss
)

type Manager struct {
	ctx            *audio.Context
	menuPlayer     *audio.Player
	gamePlayer     *audio.Player
	bossPlayer     *audio.Player
	lifeLostPlayer *audio.Player
	jumpPlayer     *audio.Player
	gameOverPlayer *audio.Player
	hey1Player     *audio.Player
	hey2Player     *audio.Player
	powerUpPlayer  *audio.Player
	mode           MusicMode
	modeSet        bool
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

	bossPlayer, err := newLoopPlayer(ctx, "game_music_boss_1.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		return nil, fmt.Errorf("boss music: %w", err)
	}
	bossPlayer.SetVolume(musicVolume)

	lifeLostPlayer, err := newOneShotPlayer(ctx, "beer_glass_hit.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		return nil, fmt.Errorf("life lost sfx: %w", err)
	}
	lifeLostPlayer.SetVolume(sfxVolume)

	jumpPlayer, err := newOneShotPlayer(ctx, "jump.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		return nil, fmt.Errorf("jump sfx: %w", err)
	}
	jumpPlayer.SetVolume(sfxVolume)

	gameOverPlayer, err := newOneShotPlayer(ctx, "laugh.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		jumpPlayer.Close()
		return nil, fmt.Errorf("game over sfx: %w", err)
	}
	gameOverPlayer.SetVolume(sfxVolume)

	hey1Player, err := newOneShotPlayer(ctx, "hey1.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		jumpPlayer.Close()
		gameOverPlayer.Close()
		return nil, fmt.Errorf("hey1 sfx: %w", err)
	}
	hey1Player.SetVolume(sfxVolume)

	hey2Player, err := newOneShotPlayer(ctx, "hey2.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		jumpPlayer.Close()
		gameOverPlayer.Close()
		hey1Player.Close()
		return nil, fmt.Errorf("hey2 sfx: %w", err)
	}
	hey2Player.SetVolume(sfxVolume)

	powerUpPlayer, err := newOneShotPlayer(ctx, "power_up.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		jumpPlayer.Close()
		gameOverPlayer.Close()
		hey1Player.Close()
		hey2Player.Close()
		return nil, fmt.Errorf("power up sfx: %w", err)
	}
	powerUpPlayer.SetVolume(sfxVolume)

	return &Manager{
		ctx:            ctx,
		menuPlayer:     menuPlayer,
		gamePlayer:     gamePlayer,
		bossPlayer:     bossPlayer,
		lifeLostPlayer: lifeLostPlayer,
		jumpPlayer:     jumpPlayer,
		gameOverPlayer: gameOverPlayer,
		hey1Player:     hey1Player,
		hey2Player:     hey2Player,
		powerUpPlayer:  powerUpPlayer,
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

func (m *Manager) PlayPowerUp() {
	if m == nil || m.powerUpPlayer == nil {
		return
	}
	_ = m.powerUpPlayer.Rewind()
	m.powerUpPlayer.Play()
}

func (m *Manager) PlayPassDad() {
	if m == nil || m.hey1Player == nil || m.hey2Player == nil {
		return
	}
	if rand.Float64() >= passDadHeyProb {
		return
	}
	player := m.hey1Player
	if rand.Intn(2) == 1 {
		player = m.hey2Player
	}
	_ = player.Rewind()
	player.Play()
}

func (m *Manager) SetMode(mode MusicMode) {
	if m == nil || (m.modeSet && m.mode == mode) {
		return
	}
	m.mode = mode
	m.modeSet = true

	m.menuPlayer.Pause()
	m.gamePlayer.Pause()
	m.bossPlayer.Pause()

	switch mode {
	case MusicMenu:
		_ = m.menuPlayer.Rewind()
		m.menuPlayer.Play()
	case MusicBoss:
		_ = m.bossPlayer.Rewind()
		m.bossPlayer.Play()
	default:
		_ = m.gamePlayer.Rewind()
		m.gamePlayer.Play()
	}
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
	if m.bossPlayer != nil {
		if e := m.bossPlayer.Close(); e != nil && err == nil {
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
	if m.hey1Player != nil {
		if e := m.hey1Player.Close(); e != nil && err == nil {
			err = e
		}
	}
	if m.hey2Player != nil {
		if e := m.hey2Player.Close(); e != nil && err == nil {
			err = e
		}
	}
	if m.powerUpPlayer != nil {
		if e := m.powerUpPlayer.Close(); e != nil && err == nil {
			err = e
		}
	}
	return err
}
