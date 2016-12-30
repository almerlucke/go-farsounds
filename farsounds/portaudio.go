package farsounds

import "github.com/gordonklaus/portaudio"

// PatchStream for port audio
type PatchStream struct {
	*portaudio.Stream
	patch     *Patch
	timestamp int64
	inlets    []*Inlet
}

// NewPatchStream new patch stream
func NewPatchStream(patch *Patch) (*PatchStream, error) {
	patchStream := new(PatchStream)
	patchStream.patch = patch

	buflen := patch.BufferLength
	numInlets := len(patch.Inlets)

	stream, err := portaudio.OpenDefaultStream(
		numInlets,
		len(patch.Outlets),
		patch.SampleRate,
		int(patch.BufferLength),
		patchStream.processAudio,
	)

	if err != nil {
		return nil, err
	}

	patchStream.inlets = make([]*Inlet, numInlets)

	for i := 0; i < numInlets; i++ {
		inlet := new(Inlet)
		inlet.Buffer = make(Buffer, buflen)
		patch.InletModules[i].Inlet = inlet
		patchStream.inlets[i] = inlet
	}

	patchStream.Stream = stream

	return patchStream, nil
}

func (stream *PatchStream) processAudio(in, out [][]float32) {
	buflen := stream.patch.BufferLength
	outlets := stream.patch.Outlets
	numInlets := len(stream.patch.Inlets)

	if numInlets > 0 {
		for i := int32(0); i < buflen; i++ {
			for j := 0; j < numInlets; j++ {
				stream.inlets[j].Buffer[i] = float64(in[j][i])
			}
		}
	}

	stream.patch.PrepareDSP()
	stream.patch.RequestDSP(stream.timestamp)

	for i := int32(0); i < buflen; i++ {
		for j := 0; j < len(outlets); j++ {
			out[j][i] = float32(outlets[j].Buffer[i])
		}
	}

	stream.timestamp += int64(buflen)
}
