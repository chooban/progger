package scan

type AppEnv struct {
	Skip  []string
	Known []string
}

func NewAppEnv() AppEnv {
	appEnv := AppEnv{
		Skip: []string{
			"Interrogation",
			"New Books",
			"Obituary",
			"Tribute",
			"Untitled",
		},
		Known: []string{
			"Anderson, Psi-Division",
			"Chimpsky's Law",
			"Counterfeit Girl",
			"Feral & Foe",
			"Lowborn High",
			"Scarlet Traces",
			"Strontium Dog",
			"Strontium Dug",
			"The Fall of Deadworld",
		},
	}

	return appEnv
}
