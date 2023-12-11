package parallel

type OptionValue struct {
	Creator CreatorWithSystem
}

type Option func(*OptionValue)

func WithCreator(c CreatorWithSystem) Option {
	return func(ov *OptionValue) {
		ov.Creator = c
	}
}
