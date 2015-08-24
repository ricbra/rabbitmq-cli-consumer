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

func (m ConfigMerger) Merge(configs []Config) (Config, error) {
	dest := Config{}
	for _, config := range configs {
		if err := mergo.MergeWithOverwrite(&dest, config); err != nil {
			return dest, fmt.Errorf("Could not merge config: %s", err.Error())
		}
	}

	return dest, nil
}
