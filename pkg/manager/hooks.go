package manager

import "time"

const (
	BeforeHook = HookName("before")
	AfterHook  = HookName("after")
)

type (
	Hooks struct {
		Before []*Hook `yaml:"before,omitempty"`
		After  []*Hook `yaml:"after,omitempty"`
	}

	HookName string

	Hook struct {
		Statements []string      `yaml:"statements"`
		Wait       time.Duration `yaml:"wait,omitempty"`
	}
)
