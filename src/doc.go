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
	varsNode, ok := rootNode.(*ast.SequenceNode)
	if !ok {
		return fmt.Errorf("expected list of env vars, got %T", rootNode)
	}
	for _, entry := range varsNode.Entries {
		mapNode, ok := entry.Value.(*ast.MappingNode)
		if !ok {
			return fmt.Errorf("expected map, got %T", rootNode)
		}
		for _, item := range mapNode.Values {
			key := item.Key.String()
			switch key {
			case "encrypted", "plain":
			case "encrypt":
				_, isStr := item.Value.(*ast.StringNode)
				if !isStr {
					return fmt.Errorf("encrypt value: expected string, got %T", item.Value)
				}
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
	}
	return nil
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
