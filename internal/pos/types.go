package pos

type UI struct {
	Title    string            `json:"title"`
	Theme    string            `json:"theme"`
	Accent   string            `json:"accent"`
	Message  string            `json:"message"`
	Toast    string            `json:"toast"`
	Busy     bool              `json:"busy"`
	Kind     string            `json:"kind"`
	Tab      string            `json:"tab,omitempty"`
	Columns  int               `json:"columns"`
	Keyboard bool              `json:"keyboard"`
	Confirm  string            `json:"confirm,omitempty"`
	Items    []Item            `json:"items,omitempty"`
	Fields   []Field           `json:"fields,omitempty"`
	Table    *Table            `json:"table,omitempty"`
	Summary  string            `json:"summary,omitempty"`
	Info     string            `json:"info,omitempty"`
	Login    bool              `json:"login,omitempty"`
	Lang     string            `json:"lang,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`
}

type Item struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Color string `json:"color,omitempty"`
}

type Field struct {
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Value       string   `json:"value"`
	Options     []string `json:"options,omitempty"`
	Placeholder string   `json:"placeholder,omitempty"`
}

type Table struct {
	Heads   []string   `json:"heads"`
	Rows    [][]string `json:"rows"`
	Actions []string   `json:"actions,omitempty"`
}

type ThemeInfo struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Accent string `json:"accent"`
}

func Themes() []ThemeInfo {
	return []ThemeInfo{
		{ID: "theme3.css", Label: "Ferari", Accent: "#ff0000"},
		{ID: "bavarska.css", Label: "Bavarska", Accent: "#669900"},
		{ID: "kajsija.css", Label: "Kajsija", Accent: "#f39c12"},
		{ID: "pinkpanter.css", Label: "Pink Panter", Accent: "#ff0080"},
	}
}

func AccentFor(name string) string {
	for _, t := range Themes() {
		if t.ID == name {
			return t.Accent
		}
	}
	return "#ff0000"
}
