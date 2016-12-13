package components

import (
	"math"

	"github.com/almerlucke/go-farsounds/farsounds"
)

const (
	// ADSRIdle idle state
	ADSRIdle = 0
	// ADSRAttack attack state
	ADSRAttack = 1
	// ADSRDecay decay state
	ADSRDecay = 2
	// ADSRSustain sustain state
	ADSRSustain = 3
	// ADSRRelease release state
	ADSRRelease = 4
)

// ADSR envelope (http://www.earlevel.com/main/2013/06/23/envelope-generators-adsr-widget/)
type ADSR struct {
	state         int
	output        float64
	attackRate    float64
	decayRate     float64
	releaseRate   float64
	attackCoef    float64
	decayCoef     float64
	releaseCoef   float64
	sustainLevel  float64
	targetRatioA  float64
	targetRatioDR float64
	attackBase    float64
	decayBase     float64
	releaseBase   float64
}

// NewADSR new ADSR
func NewADSR() *ADSR {
	adsr := new(ADSR)

	adsr.state = ADSRIdle
	adsr.sustainLevel = 1.0
	adsr.targetRatioA = 0.01
	adsr.targetRatioDR = 0.0001
	adsr.attackBase = (1.0 + adsr.targetRatioA) * (1.0 - adsr.attackCoef)
	adsr.decayBase = (adsr.sustainLevel - adsr.targetRatioDR) * (1.0 - adsr.decayCoef)
	adsr.releaseBase = -adsr.targetRatioDR * (1.0 - adsr.releaseCoef)

	return adsr
}

// Process sample please
func (adsr *ADSR) Process() float64 {
	switch adsr.state {
	case ADSRAttack:
		adsr.output = adsr.attackBase + adsr.output*adsr.attackCoef
		if adsr.output >= 1.0 {
			adsr.output = 1.0
			adsr.state = ADSRDecay
		}
	case ADSRDecay:
		adsr.output = adsr.decayBase + adsr.output*adsr.decayCoef
		if adsr.output <= adsr.sustainLevel {
			adsr.output = adsr.sustainLevel
			adsr.state = ADSRSustain
		}

		break
	case ADSRRelease:
		adsr.output = adsr.releaseBase + adsr.output*adsr.releaseCoef
		if adsr.output <= 0.0 {
			adsr.output = 0.0
			adsr.state = ADSRIdle
		}
		break
	default:
		break
	}

	return adsr.output
}

// Gate open/close
func (adsr *ADSR) Gate(gate float64) {
	if gate > 0 {
		if adsr.state == ADSRIdle || adsr.state == ADSRRelease {
			adsr.state = ADSRAttack
		}
	} else if adsr.state != ADSRIdle {
		adsr.state = ADSRRelease
	}
}

func (adsr *ADSR) calcCoef(rate float64, targetRatio float64) float64 {
	return math.Exp(-math.Log((1.0+targetRatio)/targetRatio) / rate)
}

// SetAttackRate adjust attack rate
func (adsr *ADSR) SetAttackRate(rate float64) {
	adsr.attackRate = rate
	adsr.attackCoef = adsr.calcCoef(rate, adsr.targetRatioA)
	adsr.attackBase = (1.0 + adsr.targetRatioA) * (1.0 - adsr.attackCoef)
}

// SetDecayRate adjust decay rate
func (adsr *ADSR) SetDecayRate(rate float64) {
	adsr.decayRate = rate
	adsr.decayCoef = adsr.calcCoef(rate, adsr.targetRatioDR)
	adsr.decayBase = (adsr.sustainLevel - adsr.targetRatioDR) * (1.0 - adsr.decayCoef)
}

// SetReleaseRate adjust release rate
func (adsr *ADSR) SetReleaseRate(rate float64) {
	adsr.releaseRate = rate
	adsr.releaseCoef = adsr.calcCoef(rate, adsr.targetRatioDR)
	adsr.releaseBase = -adsr.targetRatioDR * (1.0 - adsr.releaseCoef)
}

// SetSustainLevel set sustain level
func (adsr *ADSR) SetSustainLevel(level float64) {
	adsr.sustainLevel = level
	adsr.decayBase = (adsr.sustainLevel - adsr.targetRatioDR) * (1.0 - adsr.decayCoef)
}

// SetTargetRatioA adjust target ratio for attack
func (adsr *ADSR) SetTargetRatioA(targetRatio float64) {
	if targetRatio < 0.000000001 {
		targetRatio = 0.000000001
	}
	adsr.targetRatioA = targetRatio
	adsr.attackCoef = adsr.calcCoef(adsr.attackRate, adsr.targetRatioA)
	adsr.attackBase = (1.0 + adsr.targetRatioA) * (1.0 - adsr.attackCoef)
}

// SetTargetRatioDR adjust target ratio for decay and release
func (adsr *ADSR) SetTargetRatioDR(targetRatio float64) {
	if targetRatio < 0.000000001 {
		targetRatio = 0.000000001
	}
	adsr.targetRatioDR = targetRatio
	adsr.decayCoef = adsr.calcCoef(adsr.decayRate, targetRatio)
	adsr.releaseCoef = adsr.calcCoef(adsr.releaseRate, targetRatio)
	adsr.decayBase = (adsr.sustainLevel - adsr.targetRatioDR) * (1.0 - adsr.decayCoef)
	adsr.releaseBase = -adsr.targetRatioDR * (1.0 - adsr.releaseCoef)
}

// Reset ADSR
func (adsr *ADSR) Reset() {
	adsr.state = ADSRIdle
	adsr.output = 0
}

// Idle check
func (adsr *ADSR) Idle() bool {
	return adsr.state == ADSRIdle
}

/*
   ADSR module
*/

// ADSRModule ADSR module
type ADSRModule struct {
	*farsounds.BaseModule
	*ADSR
}

// NewADSRModule new ADSR module
func NewADSRModule(buflen int32, sr float64) *ADSRModule {
	adsrModule := new(ADSRModule)
	adsrModule.BaseModule = farsounds.NewBaseModule(1, 1, buflen, sr)
	adsrModule.Parent = adsrModule
	adsrModule.ADSR = NewADSR()
	return adsrModule
}

// ADSRModuleFactory creates ADSR modules
func ADSRModuleFactory(settings interface{}, buflen int32, sr float64) (farsounds.Module, error) {
	module := NewADSRModule(buflen, sr)

	module.Message(settings)

	return module, nil
}

// DSP process ADSR module
func (module *ADSRModule) DSP(timestamp int64) {
	buflen := module.GetBufferLength()

	var gateInput []float64

	output := module.Outlets[0].Buffer

	// Check if inlet is connected for gate
	if module.Inlets[0].Connections.Len() > 0 {
		gateInput = module.Inlets[0].Buffer
	}

	for i := int32(0); i < buflen; i++ {
		if gateInput != nil {
			module.Gate(gateInput[i])
		}

		output[i] = module.Process()
	}
}

// Message received
func (module *ADSRModule) Message(message farsounds.Message) {
	sr := module.GetSampleRate()

	if valueMap, ok := message.(map[string]interface{}); ok {
		if gate, ok := valueMap["gate"].(float64); ok {
			module.Gate(gate)
		}

		if targetRatioA, ok := valueMap["targetRatioA"].(float64); ok {
			module.SetTargetRatioA(targetRatioA)
		}

		if targetRatioDR, ok := valueMap["targetRatioDR"].(float64); ok {
			module.SetTargetRatioDR(targetRatioDR)
		}

		if attackRate, ok := valueMap["attackRate"].(float64); ok {
			module.SetAttackRate(attackRate * sr)
		}

		if decayRate, ok := valueMap["decayRate"].(float64); ok {
			module.SetDecayRate(decayRate * sr)
		}

		if releaseRate, ok := valueMap["releaseRate"].(float64); ok {
			module.SetReleaseRate(releaseRate * sr)
		}

		if sustainLevel, ok := valueMap["sustainLevel"].(float64); ok {
			module.SetSustainLevel(sustainLevel)
		}
	}
}
