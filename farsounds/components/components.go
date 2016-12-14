package components

import (
	"fmt"

	"github.com/almerlucke/go-farsounds/farsounds"
	"github.com/almerlucke/go-farsounds/farsounds/components/voices"
)

func init() {
	fmt.Printf("Registering components...\n")

	fmt.Printf("- register module factories\n")
	farsounds.Registry.RegisterModuleFactory("osc", OscModuleFactory)
	farsounds.Registry.RegisterModuleFactory("square", SquareModuleFactory)
	farsounds.Registry.RegisterModuleFactory("adsr", ADSRModuleFactory)
	farsounds.Registry.RegisterModuleFactory("delay", DelayModuleFactory)
	farsounds.Registry.RegisterModuleFactory("allpass", AllpassModuleFactory)

	fmt.Printf("- register poly voice factories\n\n")
	farsounds.Registry.RegisterPolyVoiceFactory("patchvoice", voices.PatchVoiceFactory, 2)
}
