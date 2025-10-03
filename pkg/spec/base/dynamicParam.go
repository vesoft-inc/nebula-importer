package specbase

type DynamicParam struct {
	Address  string `yaml:"address,omitempty"`
	User     string `yaml:"user,omitempty"`
	Password string `yaml:"password,omitempty"`
	Space    string `yaml:"space,omitempty"`
}
