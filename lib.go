package main

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

type EnvVar struct {
	Encrypted string `yaml:"encrypted"`
	Encrypt   string `yaml:"encrypt"`
	Plain     string `yaml:"plain"`
}

type Doc struct {
	Comments yaml.CommentMap
	EnvVars  []EnvVar
}

func Parse(path string) (Doc, error) {
	file, err := os.Open(path)
	if err != nil {
		return Doc{}, fmt.Errorf("open file (%s): %v", path, err)
	}
	doc := Doc{
		Comments: make(yaml.CommentMap),
		EnvVars:  make([]EnvVar, 0),
	}
	decoder := yaml.NewDecoder(
		file,
		yaml.DisallowUnknownField(),
		yaml.CommentToMap(doc.Comments),
	)
	err = decoder.Decode(&doc.EnvVars)
	if err != nil {
		return Doc{}, fmt.Errorf("parse YAML: %v", err)
	}
	return doc, nil
}
