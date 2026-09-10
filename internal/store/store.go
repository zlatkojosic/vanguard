package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type Store interface {
	PutBool(key string, value bool)
	GetBool(key string, def bool) bool
	PutFloat(key string, value float64)
	GetFloat(key string, def float64) float64
	PutString(key string, value string)
	GetString(key string, def string) string
	GetInt(key string, def int) int
	PutInt(key string, value int)
	Save() error
}

type Memory struct {
	mu   sync.RWMutex
	data map[string]any
}

func NewMemory() *Memory {
	return &Memory{data: map[string]any{}}
}

func (m *Memory) put(key string, v any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = v
}

func (m *Memory) get(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.data[key]
	return v, ok
}

func (m *Memory) PutBool(key string, value bool)          { m.put(key, value) }
func (m *Memory) PutFloat(key string, value float64)      { m.put(key, value) }
func (m *Memory) PutString(key string, value string)      { m.put(key, value) }
func (m *Memory) PutInt(key string, value int)            { m.put(key, value) }
func (m *Memory) Save() error                             { return nil }

func (m *Memory) GetBool(key string, def bool) bool {
	v, ok := m.get(key)
	if !ok {
		return def
	}
	b, ok := v.(bool)
	if !ok {
		return def
	}
	return b
}

func (m *Memory) GetFloat(key string, def float64) float64 {
	v, ok := m.get(key)
	if !ok {
		return def
	}
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	default:
		return def
	}
}

func (m *Memory) GetString(key string, def string) string {
	v, ok := m.get(key)
	if !ok {
		return def
	}
	s, ok := v.(string)
	if !ok {
		return def
	}
	return s
}

func (m *Memory) GetInt(key string, def int) int {
	v, ok := m.get(key)
	if !ok {
		return def
	}
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return def
		}
		return int(i)
	default:
		return def
	}
}

type Pref struct {
	path string
	mem  *Memory
}

func NewPref(path string) (*Pref, error) {
	p := &Pref{path: path, mem: NewMemory()}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return p, nil
		}
		return nil, err
	}
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return p, nil
	}
	p.mem.mu.Lock()
	p.mem.data = normalize(raw)
	p.mem.mu.Unlock()
	return p, nil
}

func DefaultPrefPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "vanguard", "prefs.json")
}

func (p *Pref) PutBool(key string, value bool)     { p.mem.PutBool(key, value) }
func (p *Pref) GetBool(key string, def bool) bool  { return p.mem.GetBool(key, def) }
func (p *Pref) PutFloat(key string, value float64) { p.mem.PutFloat(key, value) }
func (p *Pref) GetFloat(key string, def float64) float64 {
	return p.mem.GetFloat(key, def)
}
func (p *Pref) PutString(key string, value string)     { p.mem.PutString(key, value) }
func (p *Pref) GetString(key string, def string) string { return p.mem.GetString(key, def) }
func (p *Pref) GetInt(key string, def int) int         { return p.mem.GetInt(key, def) }
func (p *Pref) PutInt(key string, value int)           { p.mem.PutInt(key, value) }

func (p *Pref) Save() error {
	p.mem.mu.RLock()
	b, err := json.MarshalIndent(p.mem.data, "", "  ")
	p.mem.mu.RUnlock()
	if err != nil {
		return err
	}
	return os.WriteFile(p.path, b, 0o644)
}

func normalize(raw map[string]any) map[string]any {
	out := make(map[string]any, len(raw))
	for k, v := range raw {
		switch n := v.(type) {
		case float64:
			if n == float64(int(n)) {
				out[k] = int(n)
			} else {
				out[k] = n
			}
		default:
			out[k] = v
		}
	}
	return out
}

func ParseInt(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
