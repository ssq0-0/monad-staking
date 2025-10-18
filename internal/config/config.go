package config

type (
	AppConfig struct {
		Stake      Range
		Delay      Range
		Validators []string
	}

	Range struct {
		Min float32
		Max float32
	}
)
