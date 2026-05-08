package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type iniSection map[string]string
type iniFile map[string]iniSection

func readIniOptions(binaryName string) map[string]string {
	for _, candidate := range []string{filepath.Base(binaryName) + ".ini", "makecorner.ini"} {
		if _, err := os.Stat(candidate); err != nil {
			continue
		}
		parsed := parseFile(candidate)
		if parsed == nil {
			return map[string]string{}
		}
		if sec, ok := parsed["options"]; ok {
			return sec
		}
		if sec, ok := parsed[""]; ok {
			return sec
		}
	}
	return map[string]string{}
}

func parseFile(name string) iniFile {
	f, err := os.Open(name)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	section := ""
	current := iniSection{}
	sections := iniFile{}

	flush := func() {
		if len(current) > 0 {
			sections[section] = current
		}
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			flush()
			current = iniSection{}
			section = line[1 : len(line)-1]
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(value) >= 2 && strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
			value = value[1 : len(value)-1]
		}
		current[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil
	}

	flush()
	return sections
}
