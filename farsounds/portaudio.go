package farsounds

import "github.com/gordonklaus/portaudio"

// ModuleStream for port audio
type ModuleStream struct {
	*portaudio.Stream
	module    Module
	timestamp int64
}

// NewModuleStream new module stream
func NewModuleStream(module Module) (*ModuleStream, error) {
	moduleStream := new(ModuleStream)
	moduleStream.module = module

	stream, err := portaudio.OpenDefaultStream(
		0,
		len(module.GetOutlets()),
		module.GetSampleRate(),
		int(module.GetBufferLength()),
		moduleStream.processAudio,
	)

	if err != nil {
		return nil, err
	}

	moduleStream.Stream = stream

	return moduleStream, nil
}

func (stream *ModuleStream) processAudio(out [][]float32) {
	stream.module.RequestDSP(stream.timestamp)
	stream.module.DSP(stream.timestamp)

	buflen := stream.module.GetBufferLength()
	outlets := stream.module.GetOutlets()

	for i := int32(0); i < buflen; i++ {
		for j := 0; j < len(outlets); j++ {
			out[j][i] = float32(outlets[j].Buffer[i])
		}
	}

	stream.timestamp += int64(stream.module.GetBufferLength())
}
