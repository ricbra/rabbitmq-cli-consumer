package config

import (
	"fmt"

	"github.com/imdario/mergo"
)

type Merger interface {
	Merge()
}

type ConfigMerger struct {
}

func (m ConfigMerger) Merge(configs []Config) {
	dest := Config{}
	for _, config := range configs {
		if err := mergo.Merge(&dest, config); err != nil {
			fmt.Println(err)
		}
		fmt.Println(dest)

	}
}
