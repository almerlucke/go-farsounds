package voices

import "github.com/almerlucke/go-farsounds/farsounds"

const (
	patchVoiceEnvStateIdle    = 0
	patchVoiceEnvStateAttack  = 1
	patchVoiceEnvStateSustain = 2
	patchVoiceEnvStateRelease = 3
)

// PatchVoiceModule struct
type PatchVoiceModule struct {
	// Embedded base module
	*farsounds.BaseModule

	// Patch to play
	patch *farsounds.Patch

	// Simple linear fade-in/fade-out envelope
	envState   int
	env        float64
	attackInc  float64
	releaseInc float64
}

// PatchVoiceFactory factory
func PatchVoiceFactory(buflen int32, sr float64) farsounds.VoiceModule {
	patchVoiceModule := new(PatchVoiceModule)
	patchVoiceModule.BaseModule = farsounds.NewBaseModule(0, 2, buflen, sr)
	patchVoiceModule.Parent = patchVoiceModule
	return patchVoiceModule
}

func (module *PatchVoiceModule) envProcess() float64 {
	switch module.envState {
	case patchVoiceEnvStateAttack:
		module.env += module.attackInc
		if module.env >= 1.0 {
			module.env = 1.0
			module.envState = patchVoiceEnvStateSustain
		}
	case patchVoiceEnvStateRelease:
		module.env -= module.releaseInc
		if module.env <= 0.0 {
			module.env = 0.0
			module.envState = patchVoiceEnvStateIdle
		}
	}

	return module.env
}

// Cleanup module
func (module *PatchVoiceModule) Cleanup() {
	// First call base cleanup
	module.BaseModule.Cleanup()

	// Cleanup patch
	if module.patch != nil {
		module.patch.Cleanup()
		module.patch = nil
	}
}

// IsFinished for patch voice module
func (module *PatchVoiceModule) IsFinished() bool {
	return module.envState == patchVoiceEnvStateIdle
}

// NoteOff for patch voice module
func (module *PatchVoiceModule) NoteOff() {
	module.envState = patchVoiceEnvStateRelease
}

// NoteOn for patch voice module
func (module *PatchVoiceModule) NoteOn(duration float64, sr float64, settings interface{}) {
	settingsMap, ok := settings.(map[string]interface{})
	if !ok {
		return
	}

	attackDuration := 1.0
	releaseDuration := 1.0
	patchScriptPath := ""

	if path, ok := settingsMap["patch"].(string); ok {
		patchScriptPath = path
	}

	if attack, ok := settingsMap["attack"].(float64); ok {
		attackDuration = attack
	}

	if release, ok := settingsMap["release"].(float64); ok {
		releaseDuration = release
	}

	if patchScriptPath == "" {
		return
	}

	_patch, err := farsounds.PatchFactory(patchScriptPath, module.GetBufferLength(), sr)
	if err != nil {
		return
	}

	// Clean up old patch
	if module.patch != nil {
		module.patch.Cleanup()
		module.patch = nil
	}

	// Create new patch and set envelope
	module.patch = _patch.(*farsounds.Patch)
	module.envState = patchVoiceEnvStateAttack
	module.env = 0.0
	module.attackInc = 1.0 / (attackDuration * sr)
	module.releaseInc = 1.0 / (releaseDuration * sr)
}

// DSP for patch voice module
func (module *PatchVoiceModule) DSP(timestamp int64) {
	buflen := module.BufferLength

	module.patch.PrepareDSP()
	module.patch.RequestDSP(timestamp)

	for i := int32(0); i < buflen; i++ {
		env := module.envProcess()

		for j := 0; j < len(module.Outlets); j++ {
			patchBuffer := module.patch.Outlets[0].Buffer
			outBuffer := module.Outlets[j].Buffer

			if j < len(module.patch.Outlets) {
				patchBuffer = module.patch.Outlets[j].Buffer
			}

			outBuffer[i] = patchBuffer[i] * env
		}
	}
}
