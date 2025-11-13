package src

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

type EnvVar struct {
	Key   string
	Value string
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
		_, isStr := item.Value.(*ast.StringNode)
		if !isStr {
			return fmt.Errorf("value must be string, got %v", item.Value.Type())
		}
		key := item.Key.String()
		switch key {
		case "encrypted", "plain":
		case "encrypt":
			value, _ := io.ReadAll(item.Value)
			plain := crypto.NewPlainMessage(value)
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
	rootNode := doc.tree.Docs[0].Body
	varsNode, ok := rootNode.(*ast.MappingNode)
	if !ok {
		return nil, fmt.Errorf("expected list of env vars, got %T", rootNode)
	}
	vars := make([]EnvVar, 0)
	for _, entry := range varsNode.Values {
		mapNode, ok := entry.Value.(*ast.MappingNode)
		if !ok {
			return nil, fmt.Errorf("expected map, got %T", rootNode)
		}
		for _, item := range mapNode.Values {
			_, isStr := item.Value.(*ast.StringNode)
			if !isStr {
				return nil, fmt.Errorf("value must be string, got %T", item.Value)
			}
			key := item.Key.String()
			switch key {
			case "encrypted":
				encryptedBin, err := base64.StdEncoding.DecodeString(item.Value.String())
				if err != nil {
					return nil, fmt.Errorf("decode base64: %v", err)
				}
				encryptedMsg := crypto.NewPGPMessage(encryptedBin)
				decBin, err := crypto.DecryptMessageWithPassword(encryptedMsg, []byte(password))
				if err != nil {
					return nil, fmt.Errorf("decrypt field: %v", err)
				}
				vars = append(vars, EnvVar{
					Key:   entry.Key.String(),
					Value: decBin.GetString(),
				})
			case "encrypt", "plain":
				vars = append(vars, EnvVar{
					Key:   entry.Key.String(),
					Value: item.Value.String(),
				})
			default:
				return nil, fmt.Errorf("invalid key: %s", key)
			}
		}
	}
	return vars, nil
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
