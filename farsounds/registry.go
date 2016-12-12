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
	modules map[string]ModuleFactory
	tables  map[string]tables.WaveTable
	voices  map[string]*PolyVoiceFactoryEntry
}

// Registry for modules and wave tables
var Registry = &registry{
	modules: make(map[string]ModuleFactory),
	tables:  make(map[string]tables.WaveTable),
	voices:  make(map[string]*PolyVoiceFactoryEntry),
}

/*
	PolyVoice registry
*/

// RegisterPolyVoiceFactory register a poly voice factory function
func (registry *registry) RegisterPolyVoiceFactory(factoryName string, factory PolyVoiceFactory, numOutlets int) {
	registry.voices[factoryName] = &PolyVoiceFactoryEntry{
		Factory:    factory,
		NumOutlets: numOutlets,
	}
}

// GetPolyVoiceFactory get poly voice factory
func (registry *registry) GetPolyVoiceFactoryEntry(factoryName string) *PolyVoiceFactoryEntry {
	return registry.voices[factoryName]
}

/*
	Module registry
*/

// RegisterModuleFactory register a module factory function
func (registry *registry) RegisterModuleFactory(factoryName string, factory ModuleFactory) {
	registry.modules[factoryName] = factory
}

// NewModule create a new module from a factory
func (registry *registry) NewModule(factoryName string, identifier string, settings interface{}, buflen int32, sr float64) (Module, error) {
	if factory, ok := registry.modules[factoryName]; ok {
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
	registry.tables[waveTableName] = waveTable
}

// GetWaveTable get wave table from registry by name
func (registry *registry) GetWaveTable(waveTableName string) (tables.WaveTable, error) {
	if waveTable, ok := registry.tables[waveTableName]; ok {
		return waveTable, nil
	}

	return nil, fmt.Errorf("Unknown wavetable %s", waveTableName)
}
