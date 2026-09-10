package print

import (
	"strings"
	"unicode/utf8"
)

// Render produces a string receipt matching StringPrintVisitor behaviour.
func Render(t *Template, params Params, rowLen int) string {
	if rowLen <= 0 {
		rowLen = 30
	}
	v := &visitor{params: params, rowLen: rowLen}
	v.visitTemplate(t)
	v.addPreviousLine()
	return v.node.String()
}

type visitor struct {
	params      Params
	rowLen      int
	node        strings.Builder
	currentLine strings.Builder
	hasLine     bool
	align       int
}

func (v *visitor) visitTemplate(t *Template) {
	for _, c := range t.Lines {
		v.visitCell(c)
	}
}

func (v *visitor) visitCell(c Cell) {
	switch n := c.(type) {
	case Line:
		v.visitLine(n)
	case *Line:
		v.visitLine(*n)
	case PatternLine:
		v.addPreviousLine()
		v.hasLine = true
		v.currentLine.Reset()
		v.currentLine.WriteString(strings.Repeat(string(n.Pattern), v.rowLen))
	case Report:
		v.visitReport(n)
	}
}

func (v *visitor) visitLine(line Line) {
	v.addPreviousLine()
	v.align = line.Align
	if line.Condition != nil && !line.Condition(v.params) {
		return
	}
	for _, c := range line.Cells {
		v.drawCell(c)
	}
}

func (v *visitor) visitReport(r Report) {
	row := 1
	for {
		missing := false
		v.addPreviousLine()
		v.align = r.Line.Align
		col := 1
		for _, c := range r.Line.Cells {
			switch rc := c.(type) {
			case ReportColumn:
				key := "i" + itoa(row) + "." + itoa(col)
				val, ok := v.params[key]
				if !ok {
					missing = true
					break
				}
				v.currentLine.WriteString(val)
				col++
			case Space:
				v.currentLine.WriteString(strings.Repeat(" ", rc.Size))
			case StaticText:
				v.currentLine.WriteString(rc.Text)
			}
			if missing {
				break
			}
		}
		if missing {
			v.hasLine = false
			v.currentLine.Reset()
			break
		}
		row++
	}
}

func (v *visitor) drawCell(c Cell) {
	switch n := c.(type) {
	case StaticText:
		v.currentLine.WriteString(n.Text)
	case Parameter:
		val := v.params[n.Key]
		if val == "" {
			val = " "
		}
		if n.DropSpace {
			val = strings.ReplaceAll(val, " ", "")
		}
		v.currentLine.WriteString(val)
	case Space:
		v.currentLine.WriteString(strings.Repeat(" ", n.Size))
	case Custom:
		if n.Fn != nil {
			v.currentLine.WriteString(n.Fn(v.params))
		}
	case PrintImage:
		// images are ignored in the string visitor, same as Scala
	}
}

func (v *visitor) addPreviousLine() {
	if !v.hasLine && v.currentLine.Len() == 0 {
		v.hasLine = true
		return
	}
	s := v.currentLine.String()
	if v.hasLine || v.currentLine.Len() > 0 {
		switch v.align {
		case AlignCenter:
			space := (v.rowLen - utf8.RuneCountInString(s)) / 2
			if space > 0 {
				s = strings.Repeat(" ", space) + s + " "
			} else {
				s = s + " "
			}
		case AlignRight:
			space := v.rowLen - utf8.RuneCountInString(s)
			if space > 0 {
				s = strings.Repeat(" ", space) + s + " "
			} else {
				s = s + " "
			}
		default:
			if s != "" {
				s = s + " "
			} else {
				s = " "
			}
		}
		v.drawLine(s)
	}
	v.currentLine.Reset()
	v.hasLine = true
}

func (v *visitor) drawLine(s string) {
	if utf8.RuneCountInString(s) > v.rowLen && v.rowLen > 0 {
		tokens := strings.Split(s, " ")
		tokenSize := 0
		for _, token := range tokens {
			if tokenSize+len(token) > v.rowLen {
				v.node.WriteByte('\n')
				tokenSize = 0
			}
			v.node.WriteString(token + " ")
			tokenSize += len(token) + 1
		}
		v.node.WriteByte('\n')
		return
	}
	if strings.TrimSpace(s) != "" {
		v.node.WriteString(s)
		v.node.WriteByte('\n')
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
