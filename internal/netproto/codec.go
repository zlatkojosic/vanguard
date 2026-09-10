package netproto

import (
	"fmt"
	"strings"

	"nts/vanguard/internal/crypt"
)

// MapToLine encodes k=v pairs with & separators, matching Scala foldLeft order.
func MapToLine(data []KV) string {
	var b strings.Builder
	for i, kv := range data {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(kv.Key)
		b.WriteByte('=')
		b.WriteString(kv.Value)
	}
	return b.String()
}

type KV struct {
	Key, Value string
}

// LineToMap parses k=v&k2=v2. Empty values are kept as "".
func LineToMap(line string) map[string]string {
	result := make(map[string]string)
	if line == "" {
		return result
	}
	for _, item := range strings.Split(line, "&") {
		k, v, ok := strings.Cut(item, "=")
		if !ok {
			result[item] = ""
			continue
		}
		result[k] = v
	}
	return result
}

// EncodeLine appends checksum and applies SimpleCrypt.invert.
func EncodeLine(line string) string {
	cs := crypt.Checksum([]byte(line))
	withCS := fmt.Sprintf("%s&cs=%d", line, cs)
	return crypt.Invert(withCS)
}

// DecodeLine reverses invert and parses the map.
func DecodeLine(crypted string) map[string]string {
	return LineToMap(crypt.Invert(crypted))
}
