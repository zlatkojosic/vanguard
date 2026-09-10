package i18n

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

//go:embed lang_sr.properties
var langSR []byte

//go:embed lang_hr.properties
var langHR []byte

//go:embed lang_en.properties
var langEN []byte

const DefaultLang = "sr"

type Lang struct {
	ID, Label string
}

func Languages() []Lang {
	return []Lang{
		{ID: "sr", Label: "Srpski"},
		{ID: "hr", Label: "Hrvatski"},
		{ID: "en", Label: "English"},
	}
}

func Normalize(code string) string {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case "en", "eng", "english":
		return "en"
	case "hr", "hrv", "cro", "croatian":
		return "hr"
	case "sr", "srb", "bs", "bos", "bosnian", "serbian":
		return "sr"
	default:
		return DefaultLang
	}
}

type Translator struct {
	mu    sync.RWMutex
	lang  string
	packs map[string]map[string]string
}

func New() *Translator {
	t := &Translator{
		lang: DefaultLang,
		packs: map[string]map[string]string{
			"sr": load(langSR),
			"hr": load(langHR),
			"en": load(langEN),
		},
	}
	return t
}

func (t *Translator) Lang() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.lang
}

func (t *Translator) SetLang(code string) {
	t.mu.Lock()
	t.lang = Normalize(code)
	t.mu.Unlock()
}

func (t *Translator) Translate(key string) string {
	return t.T(key)
}

func (t *Translator) T(key string, args ...any) string {
	if key == "" {
		return ""
	}
	t.mu.RLock()
	lang := t.lang
	packs := t.packs
	t.mu.RUnlock()
	msg, ok := lookup(packs, lang, key)
	if !ok {
		msg = key
	}
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}

func lookup(packs map[string]map[string]string, lang, key string) (string, bool) {
	if p := packs[lang]; p != nil {
		if v, ok := p[key]; ok {
			return v, true
		}
	}
	if lang != DefaultLang {
		if p := packs[DefaultLang]; p != nil {
			if v, ok := p[key]; ok {
				return v, true
			}
		}
	}
	return "", false
}

func load(data []byte) map[string]string {
	out := map[string]string{}
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		out[strings.TrimSpace(k)] = unescape(v)
	}
	return out
}

func unescape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'u':
				if i+5 < len(s) {
					r, err := strconv.ParseInt(s[i+2:i+6], 16, 32)
					if err == nil {
						b.WriteRune(rune(r))
						i += 5
						continue
					}
				}
			case 'n':
				b.WriteByte('\n')
				i++
				continue
			default:
				b.WriteByte(s[i+1])
				i++
				continue
			}
		}
		b.WriteByte(s[i])
	}
	if !utf8.ValidString(b.String()) {
		return s
	}
	return b.String()
}
