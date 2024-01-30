package env

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
			"Strontium Dug",
		},
	}

	return appEnv
}
