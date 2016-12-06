package farsounds

import (
	"container/list"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

/*
	Patch script mapping
*/

// ScriptConnectionDescriptor for script mapping
type ScriptConnectionDescriptor struct {
	From   string
	Outlet int
	To     string
	Inlet  int
}

// ScriptModuleDescriptor for script mapping
type ScriptModuleDescriptor struct {
	Type     string
	Settings map[string]interface{}
}

// ScriptPatchSettingsDescriptor for script mapping
type ScriptPatchSettingsDescriptor struct {
	NumInlets   int
	NumOutlets  int
	Modules     map[string]interface{}
	Connections []interface{}
}

/*
   Patch inlet and outlet processors and modules creation. The patch inlet and
   outlets are used to connect the modules contained by the patch to the outside world.
*/

// InletModule is used to copy samples from outside inlet to internal inlet
type InletModule struct {
	// Ptr to base module implementation
	*BaseModule

	// Ptr to patch module inlet
	inlet *Inlet
}

// OutletModule is used to copy samples from inside outlet to outside outlet
type OutletModule struct {
	// Ptr to base module implementation
	*BaseModule

	// Ptr to patch module outlet
	outlet *Outlet
}

// NewInletModule creates a new patch inlet module
func NewInletModule(inlet *Inlet, buflen int32, sr float64) *InletModule {
	inletModule := new(InletModule)
	inletModule.BaseModule = NewBaseModule(0, 1, buflen, sr)
	inletModule.Parent = inletModule
	inletModule.inlet = inlet
	return inletModule
}

// DSP processor for patch inlet, copy patch inlet samples to module outlet
func (module *InletModule) DSP(timestamp int64) {
	// First call base module dsp
	module.BaseModule.DSP(timestamp)

	outBuffer := module.Outlets[0].Buffer
	// Copy patch inlet buffer to outlet buffer
	for i, v := range module.inlet.Buffer {
		outBuffer[i] = v
	}
}

// NewOutletModule creates a new patch outlet module
func NewOutletModule(outlet *Outlet, buflen int32, sr float64) *OutletModule {
	outletModule := new(OutletModule)
	outletModule.BaseModule = NewBaseModule(1, 0, buflen, sr)
	outletModule.Parent = outletModule
	outletModule.outlet = outlet
	return outletModule
}

// DSP processor for patch outlet, copy module inlet samples to patch outlet
func (module *OutletModule) DSP(timestamp int64) {
	// First call base module dsp
	module.BaseModule.DSP(timestamp)

	outBuffer := module.outlet.Buffer
	// Copy inlet buffer to patch outlet buffer
	for i, v := range module.Inlets[0].Buffer {
		outBuffer[i] = v
	}
}

/*
   Patch struct, processor interface methods and other methods
*/

// Patch is a container module abstracting a cluster of connected modules
// into one module
type Patch struct {
	// Ptr to base module implementation
	*BaseModule

	// Internal inlet  modules, used to copy external patch inlets to
	// internal inlets
	InletModules []*InletModule

	// Internal outlet  modules, used to copy internal outlets to
	// external outlets
	OutletModules []*OutletModule

	// List of modules in this patch
	Modules *list.List
}

// NewPatch creates a new patch module
func NewPatch(numInlets int, numOutlets int, buflen int32, sr float64) *Patch {
	patch := new(Patch)

	// Set base module
	patch.BaseModule = NewBaseModule(numInlets, numOutlets, buflen, sr)

	// Set parent ptr to self
	patch.Parent = patch

	// Create new modules list
	patch.Modules = list.New()

	// Create inlet modules
	patch.InletModules = make([]*InletModule, numInlets)

	for i := 0; i < numInlets; i++ {
		inletModule := NewInletModule(patch.Inlets[i], buflen, sr)
		// Inlet modules are identified by __inlet + number
		inletModule.SetIdentifier(fmt.Sprintf("__inlet%d", i+1))
		patch.InletModules[i] = inletModule
		patch.Modules.PushBack(inletModule)
	}

	// Create outlet modules
	patch.OutletModules = make([]*OutletModule, numOutlets)
	for i := 0; i < numOutlets; i++ {
		outletModule := NewOutletModule(patch.Outlets[i], buflen, sr)
		// Outlet modules are identified by __outlet + number
		outletModule.SetIdentifier(fmt.Sprintf("__outlet%d", i+1))
		patch.OutletModules[i] = outletModule
		patch.Modules.PushBack(outletModule)
	}

	return patch
}

// PatchFactory creates patches from settings
func PatchFactory(settings interface{}, buflen int32, sr float64) (Module, error) {
	// Create patch descriptor from raw map
	pdesc := ScriptPatchSettingsDescriptor{}

	err := mapstructure.Decode(settings, &pdesc)
	if err != nil {
		return nil, err
	}

	// Create new patch
	patch := NewPatch(pdesc.NumInlets, pdesc.NumOutlets, buflen, sr)

	// Modules lookup for making connections easier
	modules := make(map[string]Module)

	// Copy inlet modules to modules lookup so we can connect them by id
	for _, inletModule := range patch.InletModules {
		modules[inletModule.GetIdentifier()] = inletModule
	}

	// Copy outlet modules to modules lookup so we can connect them by id
	for _, outletModule := range patch.OutletModules {
		modules[outletModule.GetIdentifier()] = outletModule
	}

	// Loop through modules descriptions in map and create new modules
	for moduleIdentifier, _mdesc := range pdesc.Modules {
		// Try to get module descriptor
		mdesc := ScriptModuleDescriptor{}
		err := mapstructure.Decode(_mdesc, &mdesc)
		if err != nil {
			return nil, err
		}

		// Try to create a new module
		module, err := Registry.NewModule(mdesc.Type, moduleIdentifier, mdesc.Settings, buflen, sr)
		if err != nil {
			return nil, err
		}

		// Add to modules lookup for creating connections
		modules[moduleIdentifier] = module

		// Add module to patch
		patch.Modules.PushBack(module)
	}

	// Create connections
	for _, _cdesc := range pdesc.Connections {
		// Try to get connection descriptor
		cdesc := ScriptConnectionDescriptor{}
		err := mapstructure.Decode(_cdesc, &cdesc)
		if err != nil {
			return nil, err
		}

		from := modules[cdesc.From]
		if from == nil {
			continue
		}

		to := modules[cdesc.To]
		if from == nil {
			continue
		}

		from.Connect(cdesc.Outlet, to, cdesc.Inlet)
	}

	return patch, nil
}

// DSP processor for patch, perform DSP on internal modules
func (patch *Patch) DSP(timestamp int64) {
	// First call base module dsp
	patch.BaseModule.DSP(timestamp)

	// Prepare all modules first
	for e := patch.Modules.Front(); e != nil; e = e.Next() {
		module := e.Value.(Module)
		module.PrepareDSP()
	}

	// Loop through outlet modules and perform DSP, pulling all internally
	// connected modules. The outlet modules will copy their outputs
	// to the patch outlets
	for _, outletModule := range patch.OutletModules {
		outletModule.DSP(timestamp)
	}
}

// Cleanup all contained modules
func (patch *Patch) Cleanup() {
	for e := patch.Modules.Front(); e != nil; e = e.Next() {
		module := e.Value.(Module)
		module.Cleanup()
	}
}

// SendMessage to the patch, look at the first path component from the address,
// and see if it matches an identifier from the patch modules. If it does, check
// if the address is completely resolved, if not send the message further down
// the line, else deliver the message to the module
func (patch *Patch) SendMessage(address *Address, message Message) {
	if address.IsValid() {
		identifier := address.CurrentIdentifier()

		// Loop through all modules
		for e := patch.Modules.Front(); e != nil; e = e.Next() {
			module := e.Value.(Module)
			moduleIdentifier := module.GetIdentifier()

			// check if their identifier matches the first address
			// component identifier
			if moduleIdentifier == identifier {
				if address.IsResolved() {
					// We found the address, deliver the message
					module.Message(message)
				} else {
					// Go to next component of address
					address.Next()
					// Message is not yet on its final destination, pass it on
					module.SendMessage(address, message)
				}

				break
			}
		}
	}
}
