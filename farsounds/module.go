package farsounds

import "container/list"

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
	To    Module
	Index int
}

// Module interface
type Module interface {
	PrepareDSP()
	DSP(buflen int32, timestamp int64, samplerate int32)
	Cleanup()
	GetInlets() []*Inlet
	GetOutlets() []*Outlet
}

// BaseModule is the base module that implements all module interface methods
type BaseModule struct {
	Inlets    []*Inlet
	Outlets   []*Outlet
	Processed bool
}

// NewBaseModule creates a new basic module
func NewBaseModule(numInlets int, numOutlets int, buflen int32) *BaseModule {
	module := new(BaseModule)

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

// PrepareDSP prepares for DSP
func (module *BaseModule) PrepareDSP() {
	module.Processed = false
}

// DSP prepares inlets, calls DSP on connected modules
func (module *BaseModule) DSP(buflen int32, timestamp int64, samplerate int32) {
	// Check if we already processed for this DSP cycle, if so return
	if module.Processed {
		return
	}

	// Set to processed to prevent infinite process loops
	module.Processed = true

	// First process all inlet connections and get samples for input buffers
	for _, inlet := range module.Inlets {
		inBuffer := inlet.Buffer

		// Zero out inlet buffer
		for i := range inBuffer {
			inBuffer[i] = 0.0
		}

		// Loop through all inlet connections
		for e := inlet.Connections.Front(); e != nil; e = e.Next() {
			// Get connection
			conn := e.Value.(*Connection)

			// Call DSP of connected module
			conn.To.DSP(buflen, timestamp, samplerate)

			// Get output buffer for this connection
			outBuffer := conn.To.GetOutlets()[conn.Index].Buffer

			// Add output to input buffer
			for i, v := range outBuffer {
				inBuffer[i] += v
			}
		}
	}
}

// Cleanup disconnects inlets and outlets to break cyclic references and calls CleanupFunction
func (module *BaseModule) Cleanup() {
	// Clear references to inlets, breaking any cyclic reference so GC can reclaim objects
	for i := 0; i < len(module.Inlets); i++ {
		module.Inlets[i] = nil
	}

	// Clear references to outlets, breaking any cyclic reference so GC can reclaim objects
	for i := 0; i < len(module.Outlets); i++ {
		module.Outlets[i] = nil
	}
}

// GetInlets get inlets
func (module *BaseModule) GetInlets() []*Inlet {
	return module.Inlets
}

// GetOutlets get outlets
func (module *BaseModule) GetOutlets() []*Outlet {
	return module.Outlets
}

// IsConnected checks if there is already a connection between two modules
func IsConnected(module Module, out int, otherModule Module, in int) bool {
	outlets := module.GetOutlets()
	inlets := otherModule.GetInlets()

	if out < 0 || in < 0 || out >= len(outlets) || in >= len(inlets) {
		return false
	}

	// Loop through all connections on the out outlet to check if we already
	// connect to the in inlet of the other module
	for e := outlets[out].Connections.Front(); e != nil; e = e.Next() {
		conn := e.Value.(*Connection)

		if conn.To == otherModule && conn.Index == in {
			return true
		}
	}

	return false
}

// Connect output of this module to input of another module
func Connect(module Module, out int, otherModule Module, in int) {
	outlets := module.GetOutlets()
	inlets := otherModule.GetInlets()

	if out < 0 || in < 0 || out >= len(outlets) || in >= len(inlets) || IsConnected(module, out, otherModule, in) {
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
	_ = outlets[out].Connections.PushBack(outConn)
	_ = inlets[in].Connections.PushBack(inConn)
}

// Disconnect a module from another module
func Disconnect(module Module, out int, otherModule Module, in int) {
	outlets := module.GetOutlets()
	inlets := otherModule.GetInlets()

	if out < 0 || in < 0 || out >= len(outlets) || in >= len(inlets) {
		return
	}

	// Loop through outputs from out outlet, if we find connection, remove it
	for e := outlets[out].Connections.Front(); e != nil; e = e.Next() {
		conn := e.Value.(*Connection)

		if conn.To == otherModule && conn.Index == in {
			outlets[out].Connections.Remove(e)
			break
		}
	}

	// Loop through inputs from in inlet, if we find connection, remove it
	for e := inlets[in].Connections.Front(); e != nil; e = e.Next() {
		conn := e.Value.(*Connection)

		if conn.To == module && conn.Index == out {
			inlets[in].Connections.Remove(e)
			break
		}
	}
}
