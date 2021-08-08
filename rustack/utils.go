package rustack

import (
	"net/url"
)

type Arguments map[string]string

func Defaults() Arguments {
	return make(Arguments)
}

func (args Arguments) ToURLValues() url.Values {
	v := url.Values{}
	for key, value := range args {
		v.Set(key, value)
	}
	return v
}

func (args Arguments) merge(extraArgs []Arguments) {
	for _, extraArg := range extraArgs {
		for key, val := range extraArg {
			args[key] = val
		}
	}
}
