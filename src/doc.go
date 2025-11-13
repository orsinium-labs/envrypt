package src

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

type EnvVar struct {
	Comment string
	Key     string
	Value   string
}

type Doc struct {
	path string
	tree *ast.File
}

func Read(path string) (Doc, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return Doc{}, fmt.Errorf("read file: %v", err)
	}
	tree, err := parser.ParseBytes(file, parser.ParseComments)
	if err != nil {
		return Doc{}, fmt.Errorf("parse file: %v", err)
	}
	if len(tree.Docs) == 0 {
		return Doc{}, errors.New("empty file")
	}
	if len(tree.Docs) > 1 {
		return Doc{}, fmt.Errorf("expected one YAML doc in a file, go %d", len(tree.Docs))
	}
	doc := Doc{
		path: path,
		tree: tree,
	}
	return doc, nil
}

func (doc Doc) Encrypt(password string) error {
	if password == "" {
		return errors.New("password is empty")
	}
	if len(password) < 8 {
		return errors.New("password is too short")
	}
	rootNode := doc.tree.Docs[0].Body
	varsNode, ok := rootNode.(*ast.MappingNode)
	if !ok {
		return fmt.Errorf("expected map of env vars, got %v", rootNode.Type())
	}
	for _, entry := range varsNode.Values {
		err := doc.encryptNode(entry, password)
		if err != nil {
			name := entry.Key.String()
			return fmt.Errorf("encrypt %s: %v", name, err)
		}
	}
	return nil
}

func (doc Doc) encryptNode(entry *ast.MappingValueNode, password string) error {
	mapNode, ok := entry.Value.(*ast.MappingNode)
	if !ok {
		return fmt.Errorf("expected map, got %v", entry.Type())
	}
	for _, item := range mapNode.Values {
		valueNode, isStr := item.Value.(*ast.StringNode)
		if !isStr {
			return fmt.Errorf("value must be string, got %v", item.Value.Type())
		}
		key := item.Key.String()
		switch key {
		case "encrypted", "plain":
		case "encrypt":
			plain := crypto.NewPlainMessage([]byte(valueNode.Value))
			encBin, err := crypto.EncryptMessageWithPassword(plain, []byte(password))
			if err != nil {
				return fmt.Errorf("encrypt field: %v", err)
			}
			encStr := base64.StdEncoding.EncodeToString(encBin.Data)

			keyToken := item.Key.GetToken()
			keyToken.Value = "encrypted"
			item.Key = ast.String(keyToken)

			valToken := item.Value.GetToken()
			valToken.Value = encStr
			item.Value = ast.String(valToken)
		default:
			return fmt.Errorf("invalid key: %s", key)
		}
	}
	return nil
}

func (doc Doc) Decrypt(password string) ([]EnvVar, error) {
	if password == "" {
		return nil, errors.New("password is empty")
	}
	if len(password) < 8 {
		return nil, errors.New("password is too short")
	}
	rootNode := doc.tree.Docs[0].Body
	varsNode, ok := rootNode.(*ast.MappingNode)
	if !ok {
		return nil, fmt.Errorf("expected list of env vars, got %v", rootNode.Type())
	}
	vars := make([]EnvVar, 0)
	for _, entry := range varsNode.Values {
		envVar, err := doc.decryptEnvVar(entry, password)
		if err != nil {
			name := entry.Key.String()
			return nil, fmt.Errorf("decrypt %s: %v", name, err)
		}
		vars = append(vars, envVar)
	}
	return vars, nil
}

func (doc Doc) decryptEnvVar(entry *ast.MappingValueNode, password string) (EnvVar, error) {
	mapNode, isMap := entry.Value.(*ast.MappingNode)
	if !isMap {
		return EnvVar{}, fmt.Errorf("expected map, got %v", mapNode.Type())
	}

	// Parse node key (env var name).
	keyNode, isStr := entry.Key.(*ast.StringNode)
	if !isStr {
		return EnvVar{}, fmt.Errorf("key must be string, got %v", entry.Key.Type())
	}
	keyStr := keyNode.Value

	// Get the mapping of the env var value. One of:
	//
	//  - {"encrypt": "..."}
	//  - {"encrypted": "..."}
	//  - {"plain": "..."}
	if len(mapNode.Values) == 0 {
		return EnvVar{}, errors.New("map has no entries")
	}
	if len(mapNode.Values) > 1 {
		return EnvVar{}, fmt.Errorf("expected one entry, found %d", len(mapNode.Values))
	}
	item := mapNode.Values[0]

	// Extract (and decrypt if encrypted) the env var value.
	value, err := doc.decryptEnvVarValue(item, password)
	if err != nil {
		return EnvVar{}, fmt.Errorf("decode value: %v", err)
	}

	envVar := EnvVar{
		Comment: doc.extractComment(entry),
		Key:     keyStr,
		Value:   value,
	}
	return envVar, nil
}

func (Doc) extractComment(entry *ast.MappingValueNode) string {
	if entry.Comment == nil {
		return ""
	}
	if entry.Comment.Comments == nil {
		return ""
	}
	comments := entry.Comment.Comments
	result := make([]string, 0, len(comments))
	for _, comment := range comments {
		result = append(result, comment.String())
	}
	return strings.Join(result, "\n")
}

func (doc Doc) decryptEnvVarValue(item *ast.MappingValueNode, password string) (string, error) {
	keyNode, isStr := item.Key.(*ast.StringNode)
	if !isStr {
		return "", fmt.Errorf("key must be string, got %v", item.Value.Type())
	}
	keyStr := keyNode.Value

	valueNode, isStr := item.Value.(*ast.StringNode)
	if !isStr {
		return "", fmt.Errorf("value must be string, got %v", item.Value.Type())
	}
	valueStr := valueNode.Value

	switch keyStr {
	case "encrypted":
		encryptedBin, err := base64.StdEncoding.DecodeString(valueStr)
		if err != nil {
			return "", fmt.Errorf("decode base64: %v", err)
		}
		encryptedMsg := crypto.NewPGPMessage(encryptedBin)
		decBin, err := crypto.DecryptMessageWithPassword(encryptedMsg, []byte(password))
		if err != nil {
			return "", fmt.Errorf("decrypt field: %v", err)
		}
		return decBin.GetString(), nil
	case "encrypt", "plain":
		return valueStr, nil
	default:
		return "", fmt.Errorf("invalid key: %s", keyStr)
	}
}

func (doc Doc) Write() error {
	raw, err := io.ReadAll(doc.tree)
	if err != nil {
		return fmt.Errorf("serialize YAML: %v", err)
	}
	raw = append(raw, '\n')
	err = os.WriteFile(doc.path, raw, 0o666)
	if err != nil {
		return fmt.Errorf("write file: %v", err)
	}
	return nil
}
