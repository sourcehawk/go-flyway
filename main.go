package main

import (
	"flag"
	"log"
	"os"

	"github.com/sourcehawk/go-flyway/internal/migrator"
	"gopkg.in/yaml.v3"
)

// mergeYAML does a deep merge: later keys override earlier ones.
func mergeYAML(paths []string) (map[string]any, error) {
	out := make(map[string]any)
	for _, p := range paths {
		buf, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		var m map[string]any
		if err := yaml.Unmarshal(buf, &m); err != nil {
			return nil, err
		}
		deepMerge(out, m)
	}
	return out, nil
}

func deepMerge(dst, src map[string]any) {
	for k, v := range src {
		if vm, ok := v.(map[string]any); ok {
			if existing, found := dst[k]; found {
				if existingMap, ok := existing.(map[string]any); ok {
					deepMerge(existingMap, vm)
					dst[k] = existingMap
					continue
				}
			}
			dst[k] = vm
		} else {
			dst[k] = v
		}
	}
}

func main() {
	var configs []string
	flag.Func("config", "Path to a YAML config file (can repeat)", func(s string) error {
		configs = append(configs, s)
		return nil
	})
	flag.Parse()

	if len(configs) == 0 {
		log.Fatal("you must supply at least one --config")
	}

	merged, err := mergeYAML(configs)
	if err != nil {
		log.Fatal(err)
	}

	tmp, err := os.CreateTemp("", "merged-*.yml")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmp.Name()) //nolint:errcheck

	out, err := yaml.Marshal(merged)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := tmp.Write(out); err != nil {
		log.Fatal(err)
	}

	migrator, err := migrator.NewMigrator(tmp.Name())

	if err != nil {
		log.Fatal(err.Error())
	}

	err = migrator.Migrate()

	if err != nil {
		log.Fatal(err.Error())
	}
}
