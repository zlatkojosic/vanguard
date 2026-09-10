package crypt

// Invert matches Scala SimpleCrypt.invert: reverse bytes and XOR 7
// for ASCII printable characters in the range 32–127.
func Invert(value string) string {
	bytes := []byte(value)
	out := make([]byte, len(bytes))
	const invertByte byte = 7
	for i := 0; i < len(bytes); i++ {
		b := bytes[i]
		if b >= 32 && b <= 127 {
			out[len(bytes)-1-i] = b ^ invertByte
		} else {
			out[len(bytes)-1-i] = b
		}
	}
	return string(out)
}

// Checksum matches Scala SocketHost.calculate:
// sum of (i+1) * (b[i] & 0xFF) for each byte.
func Checksum(value []byte) int64 {
	var result int64
	for i, b := range value {
		result += int64(i+1) * int64(b&0xFF)
	}
	return result
}
