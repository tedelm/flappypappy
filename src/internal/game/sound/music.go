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

//go:embed game_music.mp3 game_music_menu.mp3 game_music_boss_1.mp3 beer_glass_hit.mp3 beer_glass_break.mp3 jump.mp3 laugh.mp3 hey1.mp3 hey2.mp3 power_up.mp3 player_win.mp3 player_win_youwin.mp3 losing_horn.mp3 ouch_1.mp3 ouch_2.mp3 wrongwithyou.mp3 footstep.mp3 laughing_run.mp3 neighbour_punch.mp3 player_punch_lose.mp3 door_open.mp3
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
	ctx                *audio.Context
	menuPlayer         *audio.Player
	gamePlayer         *audio.Player
	bossPlayer         *audio.Player
	lifeLostPlayer     *audio.Player
	glassBreakPlayer   *audio.Player
	jumpPlayer         *audio.Player
	gameOverPlayer     *audio.Player
	hey1Player         *audio.Player
	hey2Player         *audio.Player
	powerUpPlayer      *audio.Player
	playerWinPlayer    *audio.Player
	ouch1Player        *audio.Player
	ouch2Player        *audio.Player
	wrongWithYouPlayer *audio.Player
	footstepPlayer         *audio.Player
	laughingRunPlayer      *audio.Player
	neighbourPunchPlayer   *audio.Player
	playerPunchLosePlayer  *audio.Player
	doorOpenPlayer         *audio.Player
	youWinPlayer           *audio.Player
	losingHornPlayer       *audio.Player
	mode                   MusicMode
	modeSet                bool
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

	glassBreakPlayer, err := newOneShotPlayer(ctx, "beer_glass_break.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		return nil, fmt.Errorf("glass break sfx: %w", err)
	}
	glassBreakPlayer.SetVolume(sfxVolume)

	jumpPlayer, err := newOneShotPlayer(ctx, "jump.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		glassBreakPlayer.Close()
		return nil, fmt.Errorf("jump sfx: %w", err)
	}
	jumpPlayer.SetVolume(sfxVolume)

	gameOverPlayer, err := newOneShotPlayer(ctx, "laugh.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		glassBreakPlayer.Close()
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
		glassBreakPlayer.Close()
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
		glassBreakPlayer.Close()
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
		glassBreakPlayer.Close()
		jumpPlayer.Close()
		gameOverPlayer.Close()
		hey1Player.Close()
		hey2Player.Close()
		return nil, fmt.Errorf("power up sfx: %w", err)
	}
	powerUpPlayer.SetVolume(sfxVolume)

	playerWinPlayer, err := newOneShotPlayer(ctx, "player_win.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		glassBreakPlayer.Close()
		jumpPlayer.Close()
		gameOverPlayer.Close()
		hey1Player.Close()
		hey2Player.Close()
		powerUpPlayer.Close()
		return nil, fmt.Errorf("player win sfx: %w", err)
	}
	playerWinPlayer.SetVolume(sfxVolume)

	ouch1Player, err := newOneShotPlayer(ctx, "ouch_1.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		glassBreakPlayer.Close()
		jumpPlayer.Close()
		gameOverPlayer.Close()
		hey1Player.Close()
		hey2Player.Close()
		powerUpPlayer.Close()
		playerWinPlayer.Close()
		return nil, fmt.Errorf("ouch1 sfx: %w", err)
	}
	ouch1Player.SetVolume(sfxVolume)

	ouch2Player, err := newOneShotPlayer(ctx, "ouch_2.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		glassBreakPlayer.Close()
		jumpPlayer.Close()
		gameOverPlayer.Close()
		hey1Player.Close()
		hey2Player.Close()
		powerUpPlayer.Close()
		playerWinPlayer.Close()
		ouch1Player.Close()
		return nil, fmt.Errorf("ouch2 sfx: %w", err)
	}
	ouch2Player.SetVolume(sfxVolume)

	wrongWithYouPlayer, err := newOneShotPlayer(ctx, "wrongwithyou.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		glassBreakPlayer.Close()
		jumpPlayer.Close()
		gameOverPlayer.Close()
		hey1Player.Close()
		hey2Player.Close()
		powerUpPlayer.Close()
		playerWinPlayer.Close()
		ouch1Player.Close()
		ouch2Player.Close()
		return nil, fmt.Errorf("wrong with you sfx: %w", err)
	}
	wrongWithYouPlayer.SetVolume(sfxVolume)

	footstepPlayer, err := newOneShotPlayer(ctx, "footstep.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		glassBreakPlayer.Close()
		jumpPlayer.Close()
		gameOverPlayer.Close()
		hey1Player.Close()
		hey2Player.Close()
		powerUpPlayer.Close()
		playerWinPlayer.Close()
		ouch1Player.Close()
		ouch2Player.Close()
		wrongWithYouPlayer.Close()
		return nil, fmt.Errorf("footstep sfx: %w", err)
	}
	footstepPlayer.SetVolume(sfxVolume)

	laughingRunPlayer, err := newOneShotPlayer(ctx, "laughing_run.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		glassBreakPlayer.Close()
		jumpPlayer.Close()
		gameOverPlayer.Close()
		hey1Player.Close()
		hey2Player.Close()
		powerUpPlayer.Close()
		playerWinPlayer.Close()
		ouch1Player.Close()
		ouch2Player.Close()
		wrongWithYouPlayer.Close()
		footstepPlayer.Close()
		return nil, fmt.Errorf("laughing run sfx: %w", err)
	}
	laughingRunPlayer.SetVolume(sfxVolume)

	neighbourPunchPlayer, err := newOneShotPlayer(ctx, "neighbour_punch.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		glassBreakPlayer.Close()
		jumpPlayer.Close()
		gameOverPlayer.Close()
		hey1Player.Close()
		hey2Player.Close()
		powerUpPlayer.Close()
		playerWinPlayer.Close()
		ouch1Player.Close()
		ouch2Player.Close()
		wrongWithYouPlayer.Close()
		footstepPlayer.Close()
		laughingRunPlayer.Close()
		return nil, fmt.Errorf("neighbour punch sfx: %w", err)
	}
	neighbourPunchPlayer.SetVolume(sfxVolume)

	playerPunchLosePlayer, err := newOneShotPlayer(ctx, "player_punch_lose.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		glassBreakPlayer.Close()
		jumpPlayer.Close()
		gameOverPlayer.Close()
		hey1Player.Close()
		hey2Player.Close()
		powerUpPlayer.Close()
		playerWinPlayer.Close()
		ouch1Player.Close()
		ouch2Player.Close()
		wrongWithYouPlayer.Close()
		footstepPlayer.Close()
		laughingRunPlayer.Close()
		neighbourPunchPlayer.Close()
		return nil, fmt.Errorf("player punch lose sfx: %w", err)
	}
	playerPunchLosePlayer.SetVolume(sfxVolume)

	doorOpenPlayer, err := newOneShotPlayer(ctx, "door_open.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		glassBreakPlayer.Close()
		jumpPlayer.Close()
		gameOverPlayer.Close()
		hey1Player.Close()
		hey2Player.Close()
		powerUpPlayer.Close()
		playerWinPlayer.Close()
		ouch1Player.Close()
		ouch2Player.Close()
		wrongWithYouPlayer.Close()
		footstepPlayer.Close()
		laughingRunPlayer.Close()
		neighbourPunchPlayer.Close()
		playerPunchLosePlayer.Close()
		return nil, fmt.Errorf("door open sfx: %w", err)
	}
	doorOpenPlayer.SetVolume(sfxVolume)

	youWinPlayer, err := newOneShotPlayer(ctx, "player_win_youwin.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		glassBreakPlayer.Close()
		jumpPlayer.Close()
		gameOverPlayer.Close()
		hey1Player.Close()
		hey2Player.Close()
		powerUpPlayer.Close()
		playerWinPlayer.Close()
		ouch1Player.Close()
		ouch2Player.Close()
		wrongWithYouPlayer.Close()
		footstepPlayer.Close()
		laughingRunPlayer.Close()
		neighbourPunchPlayer.Close()
		playerPunchLosePlayer.Close()
		doorOpenPlayer.Close()
		return nil, fmt.Errorf("you win sfx: %w", err)
	}
	youWinPlayer.SetVolume(sfxVolume)

	losingHornPlayer, err := newOneShotPlayer(ctx, "losing_horn.mp3")
	if err != nil {
		menuPlayer.Close()
		gamePlayer.Close()
		bossPlayer.Close()
		lifeLostPlayer.Close()
		glassBreakPlayer.Close()
		jumpPlayer.Close()
		gameOverPlayer.Close()
		hey1Player.Close()
		hey2Player.Close()
		powerUpPlayer.Close()
		playerWinPlayer.Close()
		ouch1Player.Close()
		ouch2Player.Close()
		wrongWithYouPlayer.Close()
		footstepPlayer.Close()
		laughingRunPlayer.Close()
		neighbourPunchPlayer.Close()
		playerPunchLosePlayer.Close()
		doorOpenPlayer.Close()
		youWinPlayer.Close()
		return nil, fmt.Errorf("losing horn sfx: %w", err)
	}
	losingHornPlayer.SetVolume(sfxVolume)

	return &Manager{
		ctx:                   ctx,
		menuPlayer:            menuPlayer,
		gamePlayer:            gamePlayer,
		bossPlayer:            bossPlayer,
		lifeLostPlayer:        lifeLostPlayer,
		glassBreakPlayer:      glassBreakPlayer,
		jumpPlayer:            jumpPlayer,
		gameOverPlayer:        gameOverPlayer,
		hey1Player:            hey1Player,
		hey2Player:            hey2Player,
		powerUpPlayer:         powerUpPlayer,
		playerWinPlayer:       playerWinPlayer,
		ouch1Player:           ouch1Player,
		ouch2Player:           ouch2Player,
		wrongWithYouPlayer:    wrongWithYouPlayer,
		footstepPlayer:        footstepPlayer,
		laughingRunPlayer:     laughingRunPlayer,
		neighbourPunchPlayer:  neighbourPunchPlayer,
		playerPunchLosePlayer: playerPunchLosePlayer,
		doorOpenPlayer:        doorOpenPlayer,
		youWinPlayer:          youWinPlayer,
		losingHornPlayer:      losingHornPlayer,
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

func (m *Manager) PlayGlassBreak() {
	if m == nil || m.glassBreakPlayer == nil {
		return
	}
	_ = m.glassBreakPlayer.Rewind()
	m.glassBreakPlayer.Play()
}

func (m *Manager) PlayGameOver() {
	if m == nil || m.gameOverPlayer == nil {
		return
	}
	_ = m.gameOverPlayer.Rewind()
	m.gameOverPlayer.Play()
}

func (m *Manager) PlayPlayerWin() {
	if m == nil || m.playerWinPlayer == nil {
		return
	}
	_ = m.playerWinPlayer.Rewind()
	m.playerWinPlayer.Play()
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

func (m *Manager) PlayOuch() {
	if m == nil || m.ouch1Player == nil || m.ouch2Player == nil {
		return
	}
	player := m.ouch1Player
	if rand.Intn(2) == 1 {
		player = m.ouch2Player
	}
	_ = player.Rewind()
	player.Play()
}

func (m *Manager) PlayWrongWithYou() {
	if m == nil || m.wrongWithYouPlayer == nil {
		return
	}
	_ = m.wrongWithYouPlayer.Rewind()
	m.wrongWithYouPlayer.Play()
}

func (m *Manager) PlayFootstep() {
	if m == nil || m.footstepPlayer == nil {
		return
	}
	_ = m.footstepPlayer.Rewind()
	m.footstepPlayer.Play()
}

func (m *Manager) PlayLaughingRun() {
	if m == nil || m.laughingRunPlayer == nil {
		return
	}
	_ = m.laughingRunPlayer.Rewind()
	m.laughingRunPlayer.Play()
}

func (m *Manager) PlayNeighbourPunch() {
	if m == nil || m.neighbourPunchPlayer == nil {
		return
	}
	_ = m.neighbourPunchPlayer.Rewind()
	m.neighbourPunchPlayer.Play()
}

func (m *Manager) PlayPlayerPunchLose() {
	if m == nil || m.playerPunchLosePlayer == nil {
		return
	}
	_ = m.playerPunchLosePlayer.Rewind()
	m.playerPunchLosePlayer.Play()
}

func (m *Manager) PlayDoorOpen() {
	if m == nil || m.doorOpenPlayer == nil {
		return
	}
	_ = m.doorOpenPlayer.Rewind()
	m.doorOpenPlayer.Play()
}

func (m *Manager) PlayYouWin() {
	if m == nil || m.youWinPlayer == nil {
		return
	}
	_ = m.youWinPlayer.Rewind()
	m.youWinPlayer.Play()
}

func (m *Manager) PlayLosingHorn() {
	if m == nil || m.losingHornPlayer == nil {
		return
	}
	_ = m.losingHornPlayer.Rewind()
	m.losingHornPlayer.Play()
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
	if m.glassBreakPlayer != nil {
		if e := m.glassBreakPlayer.Close(); e != nil && err == nil {
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
	if m.playerWinPlayer != nil {
		if e := m.playerWinPlayer.Close(); e != nil && err == nil {
			err = e
		}
	}
	if m.ouch1Player != nil {
		if e := m.ouch1Player.Close(); e != nil && err == nil {
			err = e
		}
	}
	if m.ouch2Player != nil {
		if e := m.ouch2Player.Close(); e != nil && err == nil {
			err = e
		}
	}
	if m.wrongWithYouPlayer != nil {
		if e := m.wrongWithYouPlayer.Close(); e != nil && err == nil {
			err = e
		}
	}
	if m.footstepPlayer != nil {
		if e := m.footstepPlayer.Close(); e != nil && err == nil {
			err = e
		}
	}
	if m.laughingRunPlayer != nil {
		if e := m.laughingRunPlayer.Close(); e != nil && err == nil {
			err = e
		}
	}
	if m.neighbourPunchPlayer != nil {
		if e := m.neighbourPunchPlayer.Close(); e != nil && err == nil {
			err = e
		}
	}
	if m.playerPunchLosePlayer != nil {
		if e := m.playerPunchLosePlayer.Close(); e != nil && err == nil {
			err = e
		}
	}
	if m.doorOpenPlayer != nil {
		if e := m.doorOpenPlayer.Close(); e != nil && err == nil {
			err = e
		}
	}
	if m.youWinPlayer != nil {
		if e := m.youWinPlayer.Close(); e != nil && err == nil {
			err = e
		}
	}
	if m.losingHornPlayer != nil {
		if e := m.losingHornPlayer.Close(); e != nil && err == nil {
			err = e
		}
	}
	return err
}
