package print

const (
	AlignLeft   = 0
	AlignCenter = 1
	AlignRight  = 2
)

type Params map[string]string

type Condition func(Params) bool

type Cell interface {
	isCell()
}

type Line struct {
	Align     int
	Condition Condition
	Cells     []Cell
}

func (Line) isCell() {}

type PatternLine struct {
	Pattern rune
}

func (PatternLine) isCell() {}

type StaticText struct {
	Text      string
	Offset    int
	Bold      bool
	Align     int
	CharSize  int
	Condition Condition
}

func (StaticText) isCell() {}

type Parameter struct {
	Key       string
	Offset    int
	Bold      bool
	Align     int
	CharSize  int
	DropSpace bool
}

func (Parameter) isCell() {}

type Space struct {
	Size int
}

func (Space) isCell() {}

type Custom struct {
	Fn func(Params) string
}

func (Custom) isCell() {}

type PrintImage struct {
	Name          string
	Width, Height int
}

func (PrintImage) isCell() {}

type ReportColumn struct {
	Size  int
	Width int
	Align int
}

func (ReportColumn) isCell() {}

type Report struct {
	Columns int
	Line    *Line
	Footer  *Line
}

func (Report) isCell() {}

type Template struct {
	Lines []Cell
}

func (t *Template) Add(c Cell) {
	t.Lines = append(t.Lines, c)
}

type Builder struct {
	tmpl        *Template
	currentLine *Line
	currentCell *cellMeta
	pendingRep  *Report
}

type cellMeta struct {
	offset int
	bold   bool
	size   int
	align  int
}

func NewBuilder() *Builder {
	return &Builder{tmpl: &Template{}}
}

func (b *Builder) Template() *Template {
	b.flushLine()
	return b.tmpl
}

func (b *Builder) Add(other *Template) *Builder {
	b.flushLine()
	b.tmpl.Lines = append(b.tmpl.Lines, other.Lines...)
	return b
}

func (b *Builder) Line() *Builder {
	return b.LineAlign(AlignCenter)
}

func (b *Builder) LineAlign(align int) *Builder {
	b.flushLine()
	b.currentLine = &Line{Align: align}
	return b
}

func (b *Builder) Pattern(ch rune) *Builder {
	b.flushLine()
	b.tmpl.Add(PatternLine{Pattern: ch})
	return b
}

func (b *Builder) Text(s string) *Builder {
	b.ensureLine()
	b.currentLine.Cells = append(b.currentLine.Cells, StaticText{Text: s})
	return b
}

func (b *Builder) Param(key string) *Builder {
	b.ensureLine()
	b.currentLine.Cells = append(b.currentLine.Cells, Parameter{Key: key})
	return b
}

func (b *Builder) DropSpaceParam(key string) *Builder {
	b.ensureLine()
	b.currentLine.Cells = append(b.currentLine.Cells, Parameter{Key: key, DropSpace: true})
	return b
}

func (b *Builder) Space(n int) *Builder {
	b.ensureLine()
	b.currentLine.Cells = append(b.currentLine.Cells, Space{Size: n})
	return b
}

func (b *Builder) Custom(fn func(Params) string) *Builder {
	b.ensureLine()
	b.currentLine.Cells = append(b.currentLine.Cells, Custom{Fn: fn})
	return b
}

func (b *Builder) Image(name string, w, h int) *Builder {
	b.ensureLine()
	b.currentLine.Cells = append(b.currentLine.Cells, PrintImage{Name: name, Width: w, Height: h})
	return b
}

func (b *Builder) FontOffset(n int) *Builder { return b }
func (b *Builder) FontBold(bool) *Builder    { return b }

func (b *Builder) Report(columns int) *Builder {
	b.flushLine()
	b.pendingRep = &Report{Columns: columns, Line: &Line{Align: AlignCenter}}
	return b
}

func (b *Builder) ReportColumn() *Builder {
	if b.pendingRep != nil {
		b.pendingRep.Line.Cells = append(b.pendingRep.Line.Cells, ReportColumn{})
	}
	return b
}

func (b *Builder) ensureLine() {
	if b.currentLine == nil {
		b.currentLine = &Line{Align: AlignCenter}
	}
}

func (b *Builder) flushLine() {
	if b.pendingRep != nil {
		b.tmpl.Add(*b.pendingRep)
		b.pendingRep = nil
	}
	if b.currentLine != nil {
		b.tmpl.Add(*b.currentLine)
		b.currentLine = nil
	}
}

func FromMap(m map[string]string) Params {
	if m == nil {
		return Params{}
	}
	return Params(m)
}

func StripPromoPrefix(s string) string {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i > 0 && i+1 < len(s) && s[i] == 'c' && s[i+1] == ':' {
		rest := s[i+2:]
		return stringsTrimLeftSpace(rest)
	}
	return s
}

func stringsTrimLeftSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	return s
}

func Format4433(str string) string {
	var b []byte
	for i := 0; i < len(str); i++ {
		b = append(b, str[i])
		if i == 3 || i == 7 || i == 10 {
			b = append(b, ' ')
		}
	}
	return string(b)
}
