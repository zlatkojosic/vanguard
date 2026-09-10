package hardware

import (
	"go.bug.st/serial"
)

type Lock struct {
	release func()
}

func (l *Lock) Release() {
	if l == nil || l.release == nil {
		return
	}
	l.release()
	l.release = nil
}

func ListSerialPorts() []string {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil
	}
	return ports
}
