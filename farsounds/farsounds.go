package farsounds

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	fmt.Printf("Initialize Farsounds...\n")

	fmt.Printf("- random seed\n")
	rand.Seed(time.Now().UTC().UnixNano())

	fmt.Printf("- register module factories\n")
	Registry.RegisterModuleFactory("patch", PatchFactory)
	Registry.RegisterModuleFactory("poly", PolyVoiceModuleFactory)

	fmt.Printf("- register wave tables\n")
	Registry.RegisterWaveTable("sine", SineTable)
}
