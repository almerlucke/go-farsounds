package farsounds

import (
	"fmt"

	"github.com/almerlucke/go-farsounds/farsounds/tables"
)

// ModuleFactory is the module generator function for a factory
type ModuleFactory func(settings interface{}, buflen int32, sr float64) (Module, error)

// PolyVoiceFactoryEntry entry for the registry
type PolyVoiceFactoryEntry struct {
	Factory    PolyVoiceFactory
	NumOutlets int
}

type registry struct {
	moduleFactories map[string]ModuleFactory
	waveTables      map[string]tables.WaveTable
	voiceFactories  map[string]*PolyVoiceFactoryEntry
}

// Registry for modules and wave tables
var Registry = &registry{
	moduleFactories: make(map[string]ModuleFactory),
	waveTables:      make(map[string]tables.WaveTable),
	voiceFactories:  make(map[string]*PolyVoiceFactoryEntry),
}

/*
	PolyVoice registry
*/

// RegisterPolyVoiceFactory register a poly voice factory function
func (registry *registry) RegisterPolyVoiceFactory(factoryName string, factory PolyVoiceFactory, numOutlets int) {
	registry.voiceFactories[factoryName] = &PolyVoiceFactoryEntry{
		Factory:    factory,
		NumOutlets: numOutlets,
	}
}

// GetPolyVoiceFactory get poly voice factory
func (registry *registry) GetPolyVoiceFactoryEntry(factoryName string) *PolyVoiceFactoryEntry {
	return registry.voiceFactories[factoryName]
}

/*
	Module registry
*/

// RegisterModuleFactory register a module factory function
func (registry *registry) RegisterModuleFactory(factoryName string, factory ModuleFactory) {
	registry.moduleFactories[factoryName] = factory
}

// NewModule create a new module from a factory
func (registry *registry) NewModule(factoryName string, identifier string, settings interface{}, buflen int32, sr float64) (Module, error) {
	if factory, ok := registry.moduleFactories[factoryName]; ok {
		module, err := factory(settings, buflen, sr)
		if err != nil {
			return nil, err
		}

		module.SetIdentifier(identifier)

		return module, nil
	}

	return nil, fmt.Errorf("Unknown factory %s", factoryName)
}

/*
	Wavetable registry
*/

// RegisterWaveTable register a wave table
func (registry *registry) RegisterWaveTable(waveTableName string, waveTable tables.WaveTable) {
	registry.waveTables[waveTableName] = waveTable
}

// GetWaveTable get wave table from registry by name
func (registry *registry) GetWaveTable(waveTableName string) (tables.WaveTable, error) {
	if waveTable, ok := registry.waveTables[waveTableName]; ok {
		return waveTable, nil
	}

	return nil, fmt.Errorf("Unknown wavetable %s", waveTableName)
}
