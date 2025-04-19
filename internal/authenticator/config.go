package authenticator

type Config struct {
	URL      string `json:"url,omitempty" koanf:"url"`
	Account  string `json:"account,omitempty" koanf:"account"`
	NkeySeed string `json:"nkey_seed,omitempty" koanf:"nkey_seed"`
}
