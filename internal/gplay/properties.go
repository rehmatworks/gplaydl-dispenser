package gplay

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// DeviceConfig is a parsed Java-style .properties device profile.
type DeviceConfig map[string]string

func (d DeviceConfig) Get(key string) string { return d[key] }

func (d DeviceConfig) Int(key string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(d[key]))
	return n
}

func (d DeviceConfig) Int64(key string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(d[key]), 10, 64)
	return n
}

func (d DeviceConfig) Bool(key string) bool {
	return strings.EqualFold(strings.TrimSpace(d[key]), "true")
}

func (d DeviceConfig) List(key string) []string {
	raw := d[key]
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// LoadDeviceConfig reads resources/<name>.properties.
func LoadDeviceConfig(dir, name string) (DeviceConfig, error) {
	// Defend against path traversal since device name may come from a query param.
	if strings.ContainsAny(name, "/\\.") {
		return nil, fmt.Errorf("invalid device name %q", name)
	}
	f, err := os.Open(filepath.Join(dir, name+".properties"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg := DeviceConfig{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		eq := indexUnescaped(line, '=')
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		val = strings.ReplaceAll(val, `\:`, ":")
		val = strings.ReplaceAll(val, `\=`, "=")
		cfg[key] = val
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(cfg) == 0 {
		return nil, fmt.Errorf("device profile %q is empty", name)
	}
	return cfg, nil
}

// ListDevices returns the available device profile names in dir.
func ListDevices(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".properties") {
			names = append(names, strings.TrimSuffix(e.Name(), ".properties"))
		}
	}
	return names, nil
}

func indexUnescaped(s string, ch byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			i++
			continue
		}
		if s[i] == ch {
			return i
		}
	}
	return -1
}
