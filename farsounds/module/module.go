package module

import "container/list"

// DSPFunction is called to generate samples
type DSPFunction func(module interface{}, buflen int32, timestamp int64, samplerate int32)

// CleanupFunction is called to cleanup any resources
type CleanupFunction func(module interface{})

// Buffer is a alias for a float64 slice
type Buffer []float64

// Inlet holds connections and the input buffer
type Inlet struct {
	Buffer      Buffer
	Connections *list.List
}

// Outlet is an inlet alias
type Outlet Inlet

// Connection to an inlet/outlet index of a module
type Connection struct {
	To    *Module
	Index int
}

// Module that can be connected to other modules and that can generate samples
type Module struct {
	Inlets          []*Inlet
	Outlets         []*Outlet
	Processed       bool
	DSPFunction     DSPFunction
	CleanupFunction CleanupFunction
}

// NewModule creates a new module
func NewModule(numInlets int, numOutlets int, buflen int32, dspFunction DSPFunction, cleanupFunction CleanupFunction) *Module {
	module := new(Module)

	module.DSPFunction = dspFunction
	module.CleanupFunction = cleanupFunction

	// Create inlet and outlet slices
	module.Inlets = make([]*Inlet, numInlets)
	module.Outlets = make([]*Outlet, numOutlets)

	// Create inlets
	for i := 0; i < numInlets; i++ {
		inlet := new(Inlet)
		inlet.Buffer = make(Buffer, buflen)
		inlet.Connections = list.New()
		module.Inlets[i] = inlet
	}

	// Create outlets
	for i := 0; i < numOutlets; i++ {
		outlet := new(Outlet)
		outlet.Buffer = make(Buffer, buflen)
		outlet.Connections = list.New()
		module.Outlets[i] = outlet
	}

	return module
}

// DSP prepares inlets, calls DSP on connected modules and finally calls DSPFunction
func (module *Module) DSP(buflen int32, timestamp int64, samplerate int32) {
	if module.Processed {
		return
	}

	module.Processed = true

	// First process all inlet connections and get samples on input buffers
	for i := 0; i < len(module.Inlets); i++ {
		inlet := module.Inlets[i]
		inBuffer := inlet.Buffer

		// Zero out inlet buffer
		for j := range inBuffer {
			inBuffer[j] = 0.0
		}

		// Loop through all inlet connections
		for e := inlet.Connections.Front(); e != nil; e = e.Next() {
			// Get connection
			conn := e.Value.(*Connection)

			// Call DSP of connected module
			conn.To.DSP(buflen, timestamp, samplerate)

			// Get output buffer for this connection
			outBuffer := conn.To.Outlets[conn.Index].Buffer

			// Add output to input buffer
			for j := range inBuffer {
				inBuffer[j] += outBuffer[j]
			}
		}
	}

	if module.DSPFunction != nil {
		module.DSPFunction(module, buflen, timestamp, samplerate)
	}
}

// Cleanup disconnects inlets and outlets to break cyclic references and calls CleanupFunction
func (module *Module) Cleanup() {
	// Clear references to inlets, breaking any cyclic reference so GC can reclaim objects
	for i := 0; i < len(module.Inlets); i++ {
		module.Inlets[i] = nil
	}

	// Clear references to outlets, breaking any cyclic reference so GC can reclaim objects
	for i := 0; i < len(module.Outlets); i++ {
		module.Outlets[i] = nil
	}

	if module.CleanupFunction != nil {
		module.CleanupFunction(module)
	}
}

// Connect output of this module to input of another module
func (module *Module) Connect(out int, otherModule *Module, in int) {
	if out >= len(module.Outlets) || in >= len(module.Inlets) {
		return
	}

	// Create output and input connections
	outConn := new(Connection)
	inConn := new(Connection)

	outConn.To = otherModule
	outConn.Index = in

	inConn.To = module
	inConn.Index = out

	// Add output to outlet and input to inlet
	_ = module.Outlets[out].Connections.PushBack(outConn)
	_ = otherModule.Inlets[in].Connections.PushBack(inConn)
}
