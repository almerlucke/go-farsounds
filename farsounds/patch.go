package farsounds

import "container/list"

/*
   Patch struct, processor interface methods and other methods
*/

// Patch is a container module abstracting a cluster of connected modules
// into one module
type Patch struct {
	*BaseModule
	InletModules  []*InletModule
	OutletModules []*OutletModule
	Modules       *list.List
}

// NewPatch creates a new patch module
func NewPatch(numInlets int, numOutlets int, buflen int32) *Patch {
	patch := new(Patch)

	// Set base module
	patch.BaseModule = NewBaseModule(numInlets, numOutlets, buflen)

	// Create new modules list
	patch.Modules = list.New()

	// Create inlet modules
	patch.InletModules = make([]*InletModule, numInlets)

	for i := 0; i < numInlets; i++ {
		patch.InletModules[i] = NewInletModule(patch.Inlets[i], buflen)
		patch.Modules.PushBack(patch.InletModules[i])
	}

	// Create outlet modules
	patch.OutletModules = make([]*OutletModule, numOutlets)
	for i := 0; i < numOutlets; i++ {
		patch.OutletModules[i] = NewOutletModule(patch.Outlets[i], buflen)
		patch.Modules.PushBack(patch.OutletModules[i])
	}

	return patch
}

// DSP processor for patch, perform DSP on internal modules
func (patch *Patch) DSP(buflen int32, timestamp int64, samplerate int32) {
	// First call base module dsp
	patch.BaseModule.DSP(buflen, timestamp, samplerate)

	// Prepare all modules first
	for e := patch.Modules.Front(); e != nil; e = e.Next() {
		module := e.Value.(Module)
		module.PrepareDSP()
	}

	// Loop through outlet modules and perform DSP, pulling all internally
	// connected modules. The outlet modules will copy their outputs
	// to the patch outlets
	for _, outletModule := range patch.OutletModules {
		outletModule.DSP(buflen, timestamp, samplerate)
	}
}

// Cleanup all contained modules
func (patch *Patch) Cleanup() {
	for e := patch.Modules.Front(); e != nil; e = e.Next() {
		module := e.Value.(Module)
		module.Cleanup()
	}
}

/*
   Patch inlet and outlet processors and modules creation. The patch inlet and
   outlets are used to connect the modules contained by the patch to the outside world.
*/

// InletModule is used to copy samples from outside inlet to internal inlet
type InletModule struct {
	*BaseModule
	// Ptr to patch module inlet
	inlet *Inlet
}

// OutletModule is used to copy samples from inside outlet to outside outlet
type OutletModule struct {
	*BaseModule
	// Ptr to patch module outlet
	outlet *Outlet
}

// NewInletModule creates a new patch inlet module
func NewInletModule(inlet *Inlet, buflen int32) *InletModule {
	return &InletModule{
		BaseModule: NewBaseModule(0, 1, buflen),
		inlet:      inlet,
	}
}

// NewOutletModule creates a new patch outlet module
func NewOutletModule(outlet *Outlet, buflen int32) *OutletModule {
	return &OutletModule{
		BaseModule: NewBaseModule(1, 0, buflen),
		outlet:     outlet,
	}
}

// DSP processor for patch inlet, copy patch inlet samples to module outlet
func (module *InletModule) DSP(buflen int32, timestamp int64, samplerate int32) {
	// First call base module dsp
	module.BaseModule.DSP(buflen, timestamp, samplerate)

	outBuffer := module.Outlets[0].Buffer
	// Copy patch inlet buffer to outlet buffer
	for i, v := range module.inlet.Buffer {
		outBuffer[i] = v
	}
}

// DSP processor for patch outlet, copy module inlet samples to patch outlet
func (module *OutletModule) DSP(buflen int32, timestamp int64, samplerate int32) {
	// First call base module dsp
	module.BaseModule.DSP(buflen, timestamp, samplerate)

	outBuffer := module.outlet.Buffer
	// Copy inlet buffer to patch outlet buffer
	for i, v := range module.Inlets[0].Buffer {
		outBuffer[i] = v
	}
}
