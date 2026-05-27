package zoo

import (
	"time"

	"github.com/zkaplan/terminal-site/internal/ascii"
)

// SpeciesOrder defines the display order of species in the adopt picker.
var SpeciesOrder = []string{
	"dog", "cat", "horse", "cow", "bull",
	"pig", "piglet", "goat", "sheep",
	"chicken", "chick", "turkey",
}

// SpeciesLabel returns a human-readable label for a species key.
func SpeciesLabel(species string) string {
	labels := map[string]string{
		"dog":     "Dogs",
		"cat":     "Cats",
		"horse":   "Horses",
		"cow":     "Cows",
		"bull":    "Bulls",
		"pig":     "Pigs",
		"piglet":  "Piglets",
		"goat":    "Goats",
		"sheep":   "Sheep",
		"chicken": "Chickens",
		"chick":   "Chicks",
		"turkey":  "Turkeys",
	}
	if l, ok := labels[species]; ok {
		return l
	}
	return species
}

// Catalog is the full set of available animals, keyed by name.
var Catalog = buildCatalog()

// BySpecies returns animals grouped by species in SpeciesOrder.
func BySpecies() map[string][]*Animal {
	out := make(map[string][]*Animal)
	for _, a := range Catalog {
		out[a.Species] = append(out[a.Species], a)
	}
	return out
}

func frames(f []string, d []time.Duration) ([]string, []time.Duration) {
	return f, d
}

func buildCatalog() map[string]*Animal {
	animals := []*Animal{
		// ── Dogs ──────────────────────────────────────────────────────────────
		{
			Name: "Biscuit", Species: "dog", Breed: "beagle",
			Frames: map[State][]string{
				StateIdle:    ascii.DogBeagleIdleFrames,
				StateWalking: ascii.DogBeagleWalkFrames,
				StateEating:  ascii.DogBeagleIdleFrames, // no eat pose — reuse idle
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.DogBeagleIdleDelays,
				StateWalking: ascii.DogBeagleWalkDelays,
				StateEating:  ascii.DogBeagleIdleDelays,
			},
		},
		{
			Name: "Sunny", Species: "dog", Breed: "golden",
			Frames: map[State][]string{
				StateIdle:    ascii.DogGoldenIdleFrames,
				StateWalking: ascii.DogGoldenWalkFrames,
				StateEating:  ascii.DogGoldenIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.DogGoldenIdleDelays,
				StateWalking: ascii.DogGoldenWalkDelays,
				StateEating:  ascii.DogGoldenIdleDelays,
			},
		},
		{
			Name: "Pepper", Species: "dog", Breed: "scottie",
			Frames: map[State][]string{
				StateIdle:    ascii.DogScottieIdleFrames,
				StateWalking: ascii.DogScottieWalkFrames,
				StateEating:  ascii.DogScottieIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.DogScottieIdleDelays,
				StateWalking: ascii.DogScottieWalkDelays,
				StateEating:  ascii.DogScottieIdleDelays,
			},
		},
		{
			Name: "Bruno", Species: "dog", Breed: "brown",
			Frames: map[State][]string{
				StateIdle:    ascii.DogBrownIdleFrames,
				StateWalking: ascii.DogBrownWalkFrames,
				StateEating:  ascii.DogBrownIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.DogBrownIdleDelays,
				StateWalking: ascii.DogBrownWalkDelays,
				StateEating:  ascii.DogBrownIdleDelays,
			},
		},

		// ── Cats ──────────────────────────────────────────────────────────────
		{
			Name: "Socks", Species: "cat", Breed: "bicolor",
			Frames: map[State][]string{
				StateIdle:    ascii.CatBicolorIdleFrames,
				StateWalking: ascii.CatBicolorWalkFrames,
				StateEating:  ascii.CatBicolorIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.CatBicolorIdleDelays,
				StateWalking: ascii.CatBicolorWalkDelays,
				StateEating:  ascii.CatBicolorIdleDelays,
			},
		},
		{
			Name: "Shadow", Species: "cat", Breed: "gray",
			Frames: map[State][]string{
				StateIdle:    ascii.CatGrayIdleFrames,
				StateWalking: ascii.CatGrayWalkFrames,
				StateEating:  ascii.CatGrayIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.CatGrayIdleDelays,
				StateWalking: ascii.CatGrayWalkDelays,
				StateEating:  ascii.CatGrayIdleDelays,
			},
		},
		{
			Name: "Marmalade", Species: "cat", Breed: "tabby",
			Frames: map[State][]string{
				StateIdle:    ascii.CatTabbyIdleFrames,
				StateWalking: ascii.CatTabbyWalkFrames,
				StateEating:  ascii.CatTabbyIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.CatTabbyIdleDelays,
				StateWalking: ascii.CatTabbyWalkDelays,
				StateEating:  ascii.CatTabbyIdleDelays,
			},
		},
		{
			Name: "Luna", Species: "cat", Breed: "siamese",
			Frames: map[State][]string{
				StateIdle:    ascii.CatSiameseIdleFrames,
				StateWalking: ascii.CatSiameseWalkFrames,
				StateEating:  ascii.CatSiameseIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.CatSiameseIdleDelays,
				StateWalking: ascii.CatSiameseWalkDelays,
				StateEating:  ascii.CatSiameseIdleDelays,
			},
		},

		// ── Horses ────────────────────────────────────────────────────────────
		{
			Name: "Butter", Species: "horse", Breed: "palomino",
			Frames: map[State][]string{
				StateIdle:    ascii.HorsePalominoIdleFrames,
				StateWalking: ascii.HorsePalominoWalkFrames,
				StateEating:  ascii.HorsePalominoEatFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.HorsePalominoIdleDelays,
				StateWalking: ascii.HorsePalominoWalkDelays,
				StateEating:  ascii.HorsePalominoEatDelays,
			},
		},
		{
			Name: "Chestnut", Species: "horse", Breed: "bay",
			Frames: map[State][]string{
				StateIdle:    ascii.HorseBayIdleFrames,
				StateWalking: ascii.HorseBayWalkFrames,
				StateEating:  ascii.HorseBayEatFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.HorseBayIdleDelays,
				StateWalking: ascii.HorseBayWalkDelays,
				StateEating:  ascii.HorseBayEatDelays,
			},
		},
		{
			Name: "Storm", Species: "horse", Breed: "gray",
			Frames: map[State][]string{
				StateIdle:    ascii.HorseGrayIdleFrames,
				StateWalking: ascii.HorseGrayWalkFrames,
				StateEating:  ascii.HorseGrayEatFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.HorseGrayIdleDelays,
				StateWalking: ascii.HorseGrayWalkDelays,
				StateEating:  ascii.HorseGrayEatDelays,
			},
		},
		{
			Name: "Midnight", Species: "horse", Breed: "dark",
			Frames: map[State][]string{
				StateIdle:    ascii.HorseDarkIdleFrames,
				StateWalking: ascii.HorseDarkWalkFrames,
				StateEating:  ascii.HorseDarkEatFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.HorseDarkIdleDelays,
				StateWalking: ascii.HorseDarkWalkDelays,
				StateEating:  ascii.HorseDarkEatDelays,
			},
		},

		// ── Cows ──────────────────────────────────────────────────────────────
		{
			Name: "Clover", Species: "cow", Breed: "holstein",
			Frames: map[State][]string{
				StateIdle:    ascii.CowHolsteinIdleFrames,
				StateWalking: ascii.CowHolsteinWalkFrames,
				StateEating:  ascii.CowHolsteinEatFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.CowHolsteinIdleDelays,
				StateWalking: ascii.CowHolsteinWalkDelays,
				StateEating:  ascii.CowHolsteinEatDelays,
			},
		},

		// ── Bulls ─────────────────────────────────────────────────────────────
		{
			Name: "Angus", Species: "bull", Breed: "brown",
			Frames: map[State][]string{
				StateIdle:    ascii.BullBrownIdleFrames,
				StateWalking: ascii.BullBrownWalkFrames,
				StateEating:  ascii.BullBrownEatFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.BullBrownIdleDelays,
				StateWalking: ascii.BullBrownWalkDelays,
				StateEating:  ascii.BullBrownEatDelays,
			},
		},

		// ── Pigs ──────────────────────────────────────────────────────────────
		{
			Name: "Hamlet", Species: "pig", Breed: "pink",
			Frames: map[State][]string{
				StateIdle:    ascii.PigPinkIdleFrames,
				StateWalking: ascii.PigPinkWalkFrames,
				StateEating:  ascii.PigPinkIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.PigPinkIdleDelays,
				StateWalking: ascii.PigPinkWalkDelays,
				StateEating:  ascii.PigPinkIdleDelays,
			},
		},
		{
			Name: "Squeak", Species: "piglet", Breed: "pink",
			Frames: map[State][]string{
				StateIdle:    ascii.PigletPinkIdleFrames,
				StateWalking: ascii.PigletPinkWalkFrames,
				StateEating:  ascii.PigletPinkIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.PigletPinkIdleDelays,
				StateWalking: ascii.PigletPinkWalkDelays,
				StateEating:  ascii.PigletPinkIdleDelays,
			},
		},

		// ── Goats ─────────────────────────────────────────────────────────────
		{
			Name: "Brie", Species: "goat", Breed: "white",
			Frames: map[State][]string{
				StateIdle:    ascii.GoatWhiteIdleFrames,
				StateWalking: ascii.GoatWhiteWalkFrames,
				StateEating:  ascii.GoatWhiteIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.GoatWhiteIdleDelays,
				StateWalking: ascii.GoatWhiteWalkDelays,
				StateEating:  ascii.GoatWhiteIdleDelays,
			},
		},
		{
			Name: "Slate", Species: "goat", Breed: "dark",
			Frames: map[State][]string{
				StateIdle:    ascii.GoatDarkIdleFrames,
				StateWalking: ascii.GoatDarkWalkFrames,
				StateEating:  ascii.GoatDarkIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.GoatDarkIdleDelays,
				StateWalking: ascii.GoatDarkWalkDelays,
				StateEating:  ascii.GoatDarkIdleDelays,
			},
		},

		// ── Sheep ─────────────────────────────────────────────────────────────
		{
			Name: "Cotton", Species: "sheep", Breed: "woolly",
			Frames: map[State][]string{
				StateIdle:    ascii.SheepWoollyIdleFrames,
				StateWalking: ascii.SheepWoollyWalkFrames,
				StateEating:  ascii.SheepWoollyIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.SheepWoollyIdleDelays,
				StateWalking: ascii.SheepWoollyWalkDelays,
				StateEating:  ascii.SheepWoollyIdleDelays,
			},
		},
		{
			Name: "Cloud", Species: "sheep", Breed: "fluffy",
			Frames: map[State][]string{
				StateIdle:    ascii.SheepFluffyIdleFrames,
				StateWalking: ascii.SheepFluffyWalkFrames,
				StateEating:  ascii.SheepFluffyIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.SheepFluffyIdleDelays,
				StateWalking: ascii.SheepFluffyWalkDelays,
				StateEating:  ascii.SheepFluffyIdleDelays,
			},
		},
		{
			Name: "Ebony", Species: "sheep", Breed: "black",
			Frames: map[State][]string{
				StateIdle:    ascii.SheepBlackIdleFrames,
				StateWalking: ascii.SheepBlackWalkFrames,
				StateEating:  ascii.SheepBlackIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.SheepBlackIdleDelays,
				StateWalking: ascii.SheepBlackWalkDelays,
				StateEating:  ascii.SheepBlackIdleDelays,
			},
		},
		{
			Name: "Ash", Species: "sheep", Breed: "gray",
			Frames: map[State][]string{
				StateIdle:    ascii.SheepGrayIdleFrames,
				StateWalking: ascii.SheepGrayWalkFrames,
				StateEating:  ascii.SheepGrayIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.SheepGrayIdleDelays,
				StateWalking: ascii.SheepGrayWalkDelays,
				StateEating:  ascii.SheepGrayIdleDelays,
			},
		},

		// ── Chickens ──────────────────────────────────────────────────────────
		{
			Name: "Nugget", Species: "chicken", Breed: "brown",
			Frames: map[State][]string{
				StateIdle:    ascii.ChickenBrownIdleFrames,
				StateWalking: ascii.ChickenBrownWalkFrames,
				StateEating:  ascii.ChickenBrownIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.ChickenBrownIdleDelays,
				StateWalking: ascii.ChickenBrownWalkDelays,
				StateEating:  ascii.ChickenBrownIdleDelays,
			},
		},
		{
			Name: "Pearl", Species: "chicken", Breed: "white",
			Frames: map[State][]string{
				StateIdle:    ascii.ChickenWhiteIdleFrames,
				StateWalking: ascii.ChickenWhiteWalkFrames,
				StateEating:  ascii.ChickenWhiteIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.ChickenWhiteIdleDelays,
				StateWalking: ascii.ChickenWhiteWalkDelays,
				StateEating:  ascii.ChickenWhiteIdleDelays,
			},
		},

		// ── Chick ─────────────────────────────────────────────────────────────
		{
			Name: "Pip", Species: "chick", Breed: "yellow",
			Frames: map[State][]string{
				StateIdle:    ascii.ChickYellowIdleFrames,
				StateWalking: ascii.ChickYellowWalkFrames,
				StateEating:  ascii.ChickYellowIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.ChickYellowIdleDelays,
				StateWalking: ascii.ChickYellowWalkDelays,
				StateEating:  ascii.ChickYellowIdleDelays,
			},
		},

		// ── Turkey ────────────────────────────────────────────────────────────
		{
			Name: "Gobbles", Species: "turkey", Breed: "brown",
			Frames: map[State][]string{
				StateIdle:    ascii.TurkeyBrownIdleFrames,
				StateWalking: ascii.TurkeyBrownWalkFrames,
				StateEating:  ascii.TurkeyBrownIdleFrames,
			},
			Delays: map[State][]time.Duration{
				StateIdle:    ascii.TurkeyBrownIdleDelays,
				StateWalking: ascii.TurkeyBrownWalkDelays,
				StateEating:  ascii.TurkeyBrownIdleDelays,
			},
		},
	}

	m := make(map[string]*Animal, len(animals))
	for _, a := range animals {
		m[a.Name] = a
	}
	return m
}
