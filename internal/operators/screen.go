package operators

type Screen struct {
	Kind    string
	Title   string
	Info    string
	Summary string
	Columns int
	Items   []Item
	Fields  []Field
	Table   *Table
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

func navItem(id, title string) Item {
	return Item{ID: Encode(Action{Kind: ActNav, Target: id}), Title: title}
}

func reportItem(title, cmd, kind, label string) Item {
	return Item{ID: Encode(Action{Kind: ActReport, Target: title, Arg: cmd, Extra: kind}), Title: label}
}

func lastItem(spec, label string) Item {
	return Item{ID: Encode(Action{Kind: ActLast, Target: spec}), Title: label}
}

func buyItem(amt, label string) Item {
	return Item{ID: Encode(Action{Kind: ActBuy, Target: amt}), Title: label}
}

func codeItem(code, label string) Item {
	return Item{ID: Encode(Action{Kind: ActCode, Target: code}), Title: label}
}

func cmdItem(name, arg, label string) Item {
	return Item{ID: Encode(Action{Kind: ActCmd, Target: name, Arg: arg}), Title: label}
}

func infoScreen(id string) Screen {
	return Screen{Kind: "info", Info: id}
}

func formScreen(fields ...Field) Screen {
	return Screen{Kind: "form", Fields: fields}
}

func menuScreen(items []Item) Screen {
	return Screen{Kind: "menu", Items: items}
}
