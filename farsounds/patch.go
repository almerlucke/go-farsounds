package farsounds

import "container/list"

/*
   Patch struct, processor interface methods and other methods
*/

// Patch is a container abstracting a cluster of connected modules
// into one module
type Patch struct {
	InletModules  []*Module
	OutletModules []*Module
	Modules       *list.List
}

// NewPatchModule creates a new patch module
func NewPatchModule(numInlets int, numOutlets int, buflen int32) *Module {
	patch := new(Patch)
	module := NewModule(numInlets, numOutlets, buflen, patch)

	// Create new modules list
	patch.Modules = list.New()

	// Create inlet modules
	patch.InletModules = make([]*Module, numInlets)
	for i := 0; i < numInlets; i++ {
		patch.InletModules[i] = NewInletModule(module.Inlets[i], buflen)
		patch.Modules.PushBack(patch.InletModules[i])
	}

	// Create outlet modules
	patch.OutletModules = make([]*Module, numOutlets)
	for i := 0; i < numOutlets; i++ {
		patch.OutletModules[i] = NewOutletModule(module.Outlets[i], buflen)
		patch.Modules.PushBack(patch.OutletModules[i])
	}

	return module
}

// DSP processor for patch, perform DSP on internal modules
func (patch *Patch) DSP(module *Module, buflen int32, timestamp int64, samplerate int32) {
	// Unprocess all modules first
	for e := patch.Modules.Front(); e != nil; e = e.Next() {
		module := e.Value.(*Module)
		module.Processed = false
	}

	// Loop through outlet modules and perform DSP, pulling all internally
	// connected modules. The outlet modules will copy their outputs
	// to the patch outlets
	for _, outletModule := range patch.OutletModules {
		outletModule.DSP(buflen, timestamp, samplerate)
	}
}

// Cleanup all contained modules
func (patch *Patch) Cleanup(module *Module) {
	for e := patch.Modules.Front(); e != nil; e = e.Next() {
		module := e.Value.(*Module)
		module.Cleanup()
	}
}

/*
   Patch inlet and outlet processors and modules creation. The patch inlet and
   outlets are used to connect the modules contained by the patch to the outside world.
*/

// PatchInletProcessor is used to copy samples from outside inlet to internal inlet
type PatchInletProcessor struct {
	// Ptr to patch module inlet
	inlet *Inlet
}

// PatchOutletProcessor is used to copy samples from inside outlet to outside outlet
type PatchOutletProcessor struct {
	// Ptr to patch module outlet
	outlet *Outlet
}

// NewInletModule creates a new patch inlet module
func NewInletModule(inlet *Inlet, buflen int32) *Module {
	return NewModule(0, 1, buflen, &PatchInletProcessor{inlet: inlet})
}

// NewOutletModule creates a new patch outlet module
func NewOutletModule(outlet *Outlet, buflen int32) *Module {
	return NewModule(1, 0, buflen, &PatchOutletProcessor{outlet: outlet})
}

// DSP processor for patch inlet, copy patch inlet samples to module outlet
func (processor *PatchInletProcessor) DSP(module *Module, buflen int32, timestamp int64, samplerate int32) {
	outBuffer := module.Outlets[0].Buffer
	// Copy patch inlet buffer to outlet buffer
	for i, v := range processor.inlet.Buffer {
		outBuffer[i] = v
	}
}

// DSP processor for patch outlet, copy module inlet samples to patch outlet
func (processor *PatchOutletProcessor) DSP(module *Module, buflen int32, timestamp int64, samplerate int32) {
	outBuffer := processor.outlet.Buffer
	// Copy inlet buffer to patch outlet buffer
	for i, v := range module.Inlets[0].Buffer {
		outBuffer[i] = v
	}
}

// Cleanup stub
func (processor *PatchInletProcessor) Cleanup(module *Module) {}

// Cleanup stub
func (processor *PatchOutletProcessor) Cleanup(module *Module) {}
