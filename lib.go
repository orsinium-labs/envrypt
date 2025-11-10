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
	path     string
	comments yaml.CommentMap
	envVars  []EnvVar
}

func Read(path string) (Doc, error) {
	file, err := os.Open(path)
	if err != nil {
		return Doc{}, fmt.Errorf("open file (%s): %v", path, err)
	}
	doc := Doc{
		path:     path,
		comments: make(yaml.CommentMap),
		envVars:  make([]EnvVar, 0),
	}
	decoder := yaml.NewDecoder(
		file,
		yaml.DisallowUnknownField(),
		yaml.CommentToMap(doc.comments),
	)
	err = decoder.Decode(&doc.envVars)
	if err != nil {
		return Doc{}, fmt.Errorf("parse YAML: %v", err)
	}
	return doc, nil
}

// func (doc Doc) Write() error {
// 	file, err := os.Create(doc.path)
// 	if err != nil {
// 		return fmt.Errorf("open file (%s) for write: %v", doc.path, err)
// 	}
// 	encoder := yaml.NewEncoder(
// 		file,
// 		yaml.WithComment(doc.comments),
// 		yaml.Indent(2),
// 	)
// }
