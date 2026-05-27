package zoo

import "time"

// defaultDelay is the frame delay used for placeholder animations.
var defaultDelay = []time.Duration{200 * time.Millisecond}

// placeholder generates a simple multi-frame ASCII placeholder for an animal.
// Real frames come from internal/ascii/ once build-frames runs on actual GIFs.
func placeholder(species string, frames int) ([]string, []time.Duration) {
	// Two-frame idle wiggle using box-drawing chars.
	art := []string{
		"  /\\_/\\  \n" +
			" ( o.o ) \n" +
			"  > ^ <  \n" +
			"  [" + species[:min(3, len(species))] + "] ",
		"  /\\_/\\  \n" +
			" ( -.- ) \n" +
			"  > ^ <  \n" +
			"  [" + species[:min(3, len(species))] + "] ",
	}
	delays := []time.Duration{400 * time.Millisecond, 400 * time.Millisecond}
	if frames == 1 {
		return art[:1], delays[:1]
	}
	return art, delays
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Catalog is the full set of available animals, keyed by species name.
// Frames are placeholders — replace with generated ascii frames when GIFs are ready.
var Catalog = buildCatalog()

func buildCatalog() map[string]*Animal {
	animals := []*Animal{
		makeAnimal("Biscuit", "dog", "farm"),
		makeAnimal("Mittens", "cat", "farm"),
		makeAnimal("Thunder", "horse", "farm"),
		makeAnimal("Hamlet", "pig", "farm"),
		makeAnimal("Clover", "cow", "farm"),
		makeAnimal("Nugget", "chicken", "farm"),
		makeAnimal("Mango", "parrot", "exotic"),
		makeAnimal("Rustle", "fox", "exotic"),
		makeAnimal("Spike", "hedgehog", "exotic"),
		makeAnimal("Blip", "axolotl", "exotic"),
	}
	m := make(map[string]*Animal, len(animals))
	for _, a := range animals {
		m[a.Species] = a
	}
	return m
}

func makeAnimal(name, species, category string) *Animal {
	idle, idleD := placeholder(species, 2)
	eating, eatingD := placeholder(species, 1)
	walking, walkingD := placeholder(species, 2)
	sleeping, sleepingD := placeholder(species, 1)

	return &Animal{
		Name:     name,
		Species:  species,
		Category: category,
		Frames: map[State][]string{
			StateIdle:     idle,
			StateEating:   eating,
			StateWalking:  walking,
			StateSleeping: sleeping,
		},
		Delays: map[State][]time.Duration{
			StateIdle:     idleD,
			StateEating:   eatingD,
			StateWalking:  walkingD,
			StateSleeping: sleepingD,
		},
	}
}

// ByCategory returns animals grouped by category in a stable order.
func ByCategory() map[string][]*Animal {
	out := map[string][]*Animal{
		"farm":   {},
		"exotic": {},
	}
	for _, a := range Catalog {
		out[a.Category] = append(out[a.Category], a)
	}
	return out
}

// FarmOrder and ExoticOrder define the display order within each category.
var FarmOrder = []string{"dog", "cat", "horse", "pig", "cow", "chicken"}
var ExoticOrder = []string{"parrot", "fox", "hedgehog", "axolotl"}
