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
	Connect(out int, otherModule Module, in int)
	Disconnect(out int, otherModule Module, in int)
	IsConnected(out int, otherModule Module, in int) bool
}

// BaseModule is the base module that implements all module interface methods
type BaseModule struct {
	Parent    Module
	Inlets    []*Inlet
	Outlets   []*Outlet
	Processed bool
}

// NewBaseModule creates a new basic module
func NewBaseModule(numInlets int, numOutlets int, buflen int32) *BaseModule {
	module := new(BaseModule)

	// Set self as parent
	module.Parent = module

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
func (baseModule *BaseModule) PrepareDSP() {
	baseModule.Processed = false
}

// DSP prepares inlets, calls DSP on connected modules
func (baseModule *BaseModule) DSP(buflen int32, timestamp int64, samplerate int32) {
	// Check if we already processed for this DSP cycle, if so return
	if baseModule.Processed {
		return
	}

	// Set to processed to prevent infinite process loops
	baseModule.Processed = true

	// First process all inlet connections and get samples for input buffers
	for _, inlet := range baseModule.Inlets {
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
func (baseModule *BaseModule) Cleanup() {
	// Clear references to inlets, breaking any cyclic reference so GC can reclaim objects
	for i := 0; i < len(baseModule.Inlets); i++ {
		baseModule.Inlets[i] = nil
	}

	// Clear references to outlets, breaking any cyclic reference so GC can reclaim objects
	for i := 0; i < len(baseModule.Outlets); i++ {
		baseModule.Outlets[i] = nil
	}
}

// GetInlets get inlets
func (baseModule *BaseModule) GetInlets() []*Inlet {
	return baseModule.Inlets
}

// GetOutlets get outlets
func (baseModule *BaseModule) GetOutlets() []*Outlet {
	return baseModule.Outlets
}

// IsConnected checks if two modules are connected
func (baseModule *BaseModule) IsConnected(out int, otherModule Module, in int) bool {
	module := baseModule.Parent

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

// Connect two modules
func (baseModule *BaseModule) Connect(out int, otherModule Module, in int) {
	module := baseModule.Parent

	outlets := module.GetOutlets()
	inlets := otherModule.GetInlets()

	if out < 0 || in < 0 || out >= len(outlets) || in >= len(inlets) || module.IsConnected(out, otherModule, in) {
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

// Disconnect two modules
func (baseModule *BaseModule) Disconnect(out int, otherModule Module, in int) {
	module := baseModule.Parent

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
