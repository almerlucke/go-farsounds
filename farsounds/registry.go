package farsounds

import (
	"fmt"

	"github.com/almerlucke/go-farsounds/farsounds/tables"
)

/*
	Module registry
*/

// ModuleFactoryFunction is the module generator function for a factory
type ModuleFactoryFunction func(settings interface{}, buflen int32, samplerate int32) (Module, error)

// ModuleRegistry global module registry
var ModuleRegistry map[string]ModuleFactoryFunction

// RegisterModuleFactory register a module factory function
func RegisterModuleFactory(factoryName string, factoryFunction ModuleFactoryFunction) {
	ModuleRegistry[factoryName] = factoryFunction
}

// NewModule create a new module from a factory
func NewModule(factoryName string, identifier string, settings interface{}, buflen int32, samplerate int32) (Module, error) {
	factory := ModuleRegistry[factoryName]
	if factory == nil {
		return nil, fmt.Errorf("Unknown factory %s", factoryName)
	}

	module, err := factory(settings, buflen, samplerate)
	if err != nil {
		return nil, err
	}

	module.SetIdentifier(identifier)

	return module, nil
}

/*
	Wavetable registry
*/

// WaveTableRegistry global wave table registry
var WaveTableRegistry map[string]tables.WaveTable

// RegisterWaveTable register a wave table
func RegisterWaveTable(waveTableName string, waveTable tables.WaveTable) {
	WaveTableRegistry[waveTableName] = waveTable
}

// GetWaveTable get wave table from registry by name
func GetWaveTable(waveTableName string) (tables.WaveTable, error) {
	waveTable := WaveTableRegistry[waveTableName]
	if waveTable == nil {
		return nil, fmt.Errorf("Unknown wavetable %s", waveTableName)
	}

	return waveTable, nil
}
