package netproto

import (
	"bufio"
	"net"
	"time"
)

const DefaultTimeout = 60 * time.Second

// Host is a one-shot TCP request: one encrypted line out, one line in, then close.
type Host struct {
	Addr    string
	Timeout time.Duration
	data    []KV
}

func NewHost(ip string, port int) *Host {
	return &Host{
		Addr:    net.JoinHostPort(ip, itoa(port)),
		Timeout: DefaultTimeout,
		data:    nil,
	}
}

func (h *Host) Put(key, value string) *Host {
	h.data = append(h.data, KV{Key: key, Value: value})
	return h
}

func (h *Host) Clone() *Host {
	cp := *h
	cp.data = append([]KV(nil), h.data...)
	return &cp
}

func (h *Host) Map() map[string]string {
	out := make(map[string]string, len(h.data))
	for _, kv := range h.data {
		out[kv.Key] = kv.Value
	}
	return out
}

func (h *Host) Request() (map[string]string, error) {
	timeout := h.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	d := net.Dialer{Timeout: timeout}
	conn, err := d.Dial("tcp", h.Addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))

	line := MapToLine(h.data)
	payload := EncodeLine(line) + "\n"
	if _, err := conn.Write([]byte(payload)); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)
	resp, err := readLine(reader)
	if err != nil {
		return nil, err
	}
	return DecodeLine(resp), nil
}

func readLine(r *bufio.Reader) (string, error) {
	// Match Scala LineReader: accumulate until CR after LF.
	var data []byte
	newLine := false
	buf := make([]byte, 512)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			for i := 0; i < n; i++ {
				item := buf[i]
				if item == 10 {
					newLine = true
					continue
				}
				if item == 13 && newLine {
					return string(data), nil
				}
				if newLine {
					data = append(data, 10)
					newLine = false
				}
				data = append(data, item)
			}
		}
		if err != nil {
			if len(data) > 0 {
				return string(data), nil
			}
			return "", err
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

func IsReadTimeout(err error) bool {
	if err == nil {
		return false
	}
	ne, ok := err.(net.Error)
	return ok && ne.Timeout()
}
