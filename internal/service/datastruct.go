package service

type (
	RunParams struct {
		Stake           Range
		Delay           Range
		Validators      []uint8
		ContractAddress string
	}

	Range struct {
		Min float32
		Max float32
	}
)
