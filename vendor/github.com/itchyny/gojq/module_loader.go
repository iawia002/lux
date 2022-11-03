package gojq

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type moduleLoader struct {
	paths []string
}

// NewModuleLoader creates a new ModuleLoader reading local modules in the paths.
func NewModuleLoader(paths []string) ModuleLoader {
	return &moduleLoader{expandHomeDir(paths)}
}

func (l *moduleLoader) LoadInitModules() ([]*Query, error) {
	var qs []*Query
	for _, path := range l.paths {
		if filepath.Base(path) != ".jq" {
			continue
		}
		fi, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		if fi.IsDir() {
			continue
		}
		cnt, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		q, err := parseModule(path, string(cnt))
		if err != nil {
			return nil, &queryParseError{"query in module", path, string(cnt), err}
		}
		qs = append(qs, q)
	}
	return qs, nil
}

func (l *moduleLoader) LoadModuleWithMeta(name string, meta map[string]interface{}) (*Query, error) {
	path, err := l.lookupModule(name, ".jq", meta)
	if err != nil {
		return nil, err
	}
	cnt, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	q, err := parseModule(path, string(cnt))
	if err != nil {
		return nil, &queryParseError{"query in module", path, string(cnt), err}
	}
	return q, nil
}

func (l *moduleLoader) LoadJSONWithMeta(name string, meta map[string]interface{}) (interface{}, error) {
	path, err := l.lookupModule(name, ".json", meta)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var vals []interface{}
	dec := json.NewDecoder(f)
	dec.UseNumber()
	for {
		var val interface{}
		if err := dec.Decode(&val); err != nil {
			if err == io.EOF {
				break
			}
			if _, err := f.Seek(0, io.SeekStart); err != nil {
				return nil, err
			}
			cnt, er := io.ReadAll(f)
			if er != nil {
				return nil, er
			}
			return nil, &jsonParseError{path, string(cnt), err}
		}
		vals = append(vals, val)
	}
	return vals, nil
}

func (l *moduleLoader) lookupModule(name, extension string, meta map[string]interface{}) (string, error) {
	paths := l.paths
	if path := searchPath(meta); path != "" {
		paths = append([]string{path}, paths...)
	}
	for _, base := range paths {
		path := filepath.Clean(filepath.Join(base, name+extension))
		if _, err := os.Stat(path); err == nil {
			return path, err
		}
		path = filepath.Clean(filepath.Join(base, name, filepath.Base(name)+extension))
		if _, err := os.Stat(path); err == nil {
			return path, err
		}
	}
	return "", fmt.Errorf("module not found: %q", name)
}

// This is a dirty hack to implement the "search" field.
func parseModule(path, cnt string) (*Query, error) {
	q, err := Parse(cnt)
	if err != nil {
		return nil, err
	}
	for _, i := range q.Imports {
		if i.Meta == nil {
			continue
		}
		i.Meta.KeyVals = append(
			i.Meta.KeyVals,
			&ConstObjectKeyVal{
				Key: "$$path",
				Val: &ConstTerm{Str: path},
			},
		)
	}
	return q, nil
}

func searchPath(meta map[string]interface{}) string {
	x, ok := meta["search"]
	if !ok {
		return ""
	}
	s, ok := x.(string)
	if !ok {
		return ""
	}
	if filepath.IsAbs(s) {
		return s
	}
	if strings.HasPrefix(s, "~") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			return filepath.Join(homeDir, s[1:])
		}
	}
	var path string
	if x, ok := meta["$$path"]; ok {
		path, _ = x.(string)
	}
	if path == "" {
		return s
	}
	return filepath.Join(filepath.Dir(path), s)
}

func expandHomeDir(paths []string) []string {
	var homeDir string
	var err error
	for i, path := range paths {
		if strings.HasPrefix(path, "~") {
			if homeDir == "" && err == nil {
				homeDir, err = os.UserHomeDir()
			}
			if homeDir != "" {
				paths[i] = filepath.Join(homeDir, path[1:])
			}
		}
	}
	return paths
}
