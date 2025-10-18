package config

type (
	AppConfig struct {
		Stake           Range    `yaml:"stake"`
		Delay           Range    `yaml:"delay"`
		Validators      []string `yaml:"validators"`
		PrivateKeysFile string   `yaml:"privateKeysFile"`
		RPCString       string   `yaml:"rpc"`
	}

	Range struct {
		Min float32 `yaml:"min"`
		Max float32 `yaml:"max"`
	}
)
