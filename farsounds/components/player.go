package components

import (
	"math"
	"path/filepath"

	"github.com/almerlucke/go-farsounds/farsounds"
)

// PlayerModule file player
type PlayerModule struct {
	// Base module
	*farsounds.BaseModule

	// Sound file buffer
	buffer *farsounds.SoundFileBuffer

	// Params
	position float64
	inc      float64
	speed    float64
	startPos float64
	endPos   float64
	repeat   bool
	stopped  bool
}

// NewPlayerModule module
func NewPlayerModule(filePath string, speed float64, repeat bool, startPos float64, endPos float64, buflen int32, sr float64) (*PlayerModule, error) {
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	buffer := farsounds.Registry.GetSoundFileBuffer(absFilePath)
	if buffer == nil {
		newBuffer, err := farsounds.NewSoundFileBuffer(absFilePath)
		if err != nil {
			return nil, err
		}

		farsounds.Registry.RegisterSoundFileBuffer(absFilePath, newBuffer)
		buffer = newBuffer
	}

	player := new(PlayerModule)
	player.BaseModule = farsounds.NewBaseModule(0, len(buffer.Channels), buflen, sr)
	player.Parent = player
	player.buffer = buffer

	if startPos > 0.0 {
		player.startPos = startPos * buffer.SampleRate
	}

	if endPos > 0.0 {
		player.endPos = endPos * buffer.SampleRate
	} else {
		player.endPos = float64(buffer.NumFrames)
	}

	player.inc = speed * buffer.SampleRate / sr
	player.position = player.startPos

	if speed < 0.0 {
		player.position = player.endPos
	}

	player.speed = speed
	player.repeat = repeat
	player.stopped = false

	return player, nil
}

// PlayerModuleFactory module factory
func PlayerModuleFactory(settings interface{}, buflen int32, sr float64) (farsounds.Module, error) {
	settingsMap := settings.(map[string]interface{})
	filePath := ""
	speed := 1.0
	repeat := true
	startPos := 0.0
	endPos := 0.0

	if _filePath, ok := settingsMap["file"].(string); ok {
		filePath = _filePath
	}

	if _speed, ok := settingsMap["speed"].(float64); ok {
		speed = _speed
	}

	if _repeat, ok := settingsMap["repeat"].(bool); ok {
		repeat = _repeat
	}

	if _startPos, ok := settingsMap["start"].(float64); ok {
		startPos = _startPos
	}

	if _endPos, ok := settingsMap["end"].(float64); ok {
		endPos = _endPos
	}

	return NewPlayerModule(filePath, speed, repeat, startPos, endPos, buflen, sr)
}

// DSP perform
func (player *PlayerModule) DSP(timestamp int64) {
	buflen := player.BufferLength
	outlets := player.Outlets
	buffer := player.buffer

	posRange := player.endPos - player.startPos

	for i := int32(0); i < buflen; i++ {
		for j := 0; j < len(buffer.Channels); j++ {
			outlets[j].Buffer[i] = 0.0
		}

		if !player.stopped {
			pos := player.position

			firstIndexF, fraction := math.Modf(pos)
			firstIndex := int64(firstIndexF)
			secondIndex := int64(firstIndex + 1)

			if firstIndex >= buffer.NumFrames {
				firstIndex = buffer.NumFrames - 1
			}

			if secondIndex >= buffer.NumFrames {
				secondIndex = buffer.NumFrames - 1
			}

			for j := 0; j < len(buffer.Channels); j++ {
				chanBuffer := buffer.Channels[j]
				v1 := chanBuffer[firstIndex]
				v2 := chanBuffer[secondIndex]
				outlets[j].Buffer[i] = v1 + (v2-v1)*fraction
			}

			pos += player.inc

			if pos >= player.endPos {
				if !player.repeat {
					player.stopped = true
				} else {
					for pos >= player.endPos {
						pos -= posRange
					}
				}
			}

			if pos < player.startPos {
				if !player.repeat {
					player.stopped = true
				} else {
					for pos < player.startPos {
						pos += posRange
					}
				}
			}

			player.position = pos
		}
	}
}
