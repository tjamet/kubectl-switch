package xtesting

import (
	"os"
	"path/filepath"
	"strings"
)

func parseRaw(raw string) (key string, value string) {
	kv := strings.Split(raw, "=")
	if len(kv) < 2 {
		return kv[0], ""
	}
	return kv[0], strings.Join(kv[1:], "=")
}

func NoEnv(pattern string) func() {
	oldEnv := map[string]string{}
	for _, raw := range os.Environ() {
		key, value := parseRaw(raw)
		matched, err := filepath.Match(pattern, key)
		if err != nil {
			panic(err)
		}
		if matched {
			oldEnv[key] = value
			err = os.Unsetenv(key)
			if err != nil {
				panic(err)
			}
		}
	}
	return func() {
		for _, raw := range os.Environ() {
			key, _ := parseRaw(raw)
			matched, err := filepath.Match(pattern, key)
			if err != nil {
				panic(err)
			}
			if matched {
				err = os.Unsetenv(key)
				if err != nil {
					panic(err)
				}
			}
		}
		for key, val := range oldEnv {
			err := os.Setenv(key, val)
			if err != nil {
				panic(err)
			}
		}
	}
}

func InEnv(key, value string) func() {
	f := NoEnv(key)
	err := os.Setenv(key, value)
	if err != nil {
		panic(err)
	}
	return f
}
