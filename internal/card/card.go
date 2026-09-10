package card

import (
	"strings"
	"unicode"
)

type Event struct {
	CardNumber string
}

type Reader interface {
	HandleRune(r rune) (ev *Event, consume bool)
	HandleEnter() (ev *Event, consume bool)
}

type EngReader struct {
	buf     strings.Builder
	consume bool
}

func NewEng() *EngReader { return &EngReader{} }

func isCardStart(r rune) bool {
	s := string(r)
	return s == ";" || s == "è" || s == "č" || s == "ч" || s == "Č"
}

func (e *EngReader) HandleRune(r rune) (*Event, bool) {
	if isCardStart(r) {
		e.consume = true
		e.buf.Reset()
		return nil, true
	}
	if e.consume && r != '/' {
		e.buf.WriteRune(r)
		return nil, true
	}
	return nil, e.consume
}

func (e *EngReader) HandleEnter() (*Event, bool) {
	if !e.consume {
		return nil, false
	}
	e.consume = false
	v := e.buf.String()
	v = strings.ReplaceAll(v, ";", "")
	v = strings.ReplaceAll(v, "/", "")
	v = strings.ReplaceAll(v, "-", "")
	e.buf.Reset()
	return &Event{CardNumber: v}, true
}

type LampicaReader struct {
	state int // 0=wait5, 1=wait6, 2=wait7, 3=rest
	buf   strings.Builder
}

func NewLampica() *LampicaReader { return &LampicaReader{} }

func (l *LampicaReader) HandleEnter() (*Event, bool) { return nil, false }

func (l *LampicaReader) HandleRune(r rune) (*Event, bool) {
	s := string(r)
	switch l.state {
	case 0:
		if s == "5" {
			l.state = 1
		}
	case 1:
		if s == "6" {
			l.state = 2
		} else {
			l.state = 0
		}
	case 2:
		if s == "7" {
			l.state = 3
			l.buf.Reset()
		} else {
			l.state = 0
		}
	case 3:
		if unicode.IsDigit(r) {
			l.buf.WriteRune(r)
			if l.buf.Len() == 12 {
				rest := l.buf.String()
				l.state = 0
				l.buf.Reset()
				if ValidateLuhn(rest) {
					return &Event{CardNumber: "567" + rest}, true
				}
			}
			return nil, true
		}
		l.state = 0
		l.buf.Reset()
	}
	return nil, l.state > 0
}

func addDigit(digit byte, double bool) int {
	n := int(digit - '0')
	if !double {
		return n
	}
	n = n << 1
	if n < 10 {
		return n
	}
	return (n / 10) + (n % 10)
}

func CalculateCheckSum(card string) int {
	sum := 0
	for i := 0; i < len(card); i++ {
		sum += addDigit(card[len(card)-1-i], (i&1) == 0)
	}
	sum = sum % 10
	if sum != 0 {
		sum = 10 - sum%10
	}
	return sum
}

func ValidateLuhn(card string, checkFor ...int) bool {
	checkLen := len(card)
	if len(checkFor) > 0 && checkFor[0] != 0 {
		checkLen = checkFor[0]
	}
	sum := 0
	for i := 1; i <= len(card)-1; i++ {
		idx := checkLen - i - 1
		if idx < 0 || idx >= len(card) {
			return false
		}
		sum += addDigit(card[idx], (i&1) != 0)
	}
	return sum%10 == 0
}

type Mux struct {
	readers []Reader
}

func NewMux(rs ...Reader) *Mux { return &Mux{readers: rs} }

func (m *Mux) HandleRune(r rune) *Event {
	for _, rd := range m.readers {
		if ev, _ := rd.HandleRune(r); ev != nil {
			return ev
		}
	}
	return nil
}

func (m *Mux) HandleEnter() *Event {
	for _, rd := range m.readers {
		if ev, _ := rd.HandleEnter(); ev != nil {
			return ev
		}
	}
	return nil
}
