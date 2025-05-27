package authenticator

type User struct {
	Password   string `json:"password,omitempty"   koanf:"password"`
	Account    string `json:"account,omitempty"    koanf:"account"`
	Privileged bool   `json:"privileged,omitempty" koanf:"privileged"`
}

type Config struct {
	URL      string          `json:"url,omitempty"       koanf:"url"`
	NkeySeed string          `json:"nkey_seed,omitempty" koanf:"nkey_seed"`
	Users    map[string]User `json:"users,omitempty"     koanf:"users"`
}
