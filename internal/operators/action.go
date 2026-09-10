package operators

import "strings"

const (
	ActNav    = "nav"
	ActTheme  = "theme"
	ActLang   = "lang"
	ActReport = "report"
	ActLast   = "last"
	ActBuy    = "buy"
	ActCode   = "code"
	ActCmd    = "cmd"
)

type Action struct {
	Kind   string
	Target string
	Arg    string
	Extra  string
}

func Encode(a Action) string {
	switch a.Kind {
	case ActNav:
		return "go:" + a.Target
	case ActTheme:
		return "theme:" + a.Target
	case ActLang:
		return "lang:" + a.Target
	case ActReport:
		id := "report:" + a.Target
		if a.Arg != "" {
			id += ":" + a.Arg
		}
		if a.Extra != "" {
			id += ":" + a.Extra
		}
		return id
	case ActLast:
		return "last:" + a.Target
	case ActBuy:
		return "amt:" + a.Target
	case ActCode:
		return "bin:" + a.Target
	case ActCmd:
		switch a.Target {
		case "pending":
			return "acPending"
		case "copy":
			return "acLast"
		case "reportShow":
			return "reportShow"
		case "reportPrint":
			return "reportPrint"
		case "lampLast":
			return "lampLast:" + a.Arg
		case "storno":
			return "xbonStorno:" + a.Arg
		default:
			if a.Arg != "" {
				return "cmd:" + a.Target + ":" + a.Arg
			}
			return "cmd:" + a.Target
		}
	default:
		return a.Target
	}
}

func Decode(id string) Action {
	switch {
	case id == "":
		return Action{}
	case strings.HasPrefix(id, "go:"):
		return Action{Kind: ActNav, Target: strings.TrimPrefix(id, "go:")}
	case strings.HasPrefix(id, "nav:"):
		return Action{Kind: ActNav, Target: strings.TrimPrefix(id, "nav:")}
	case strings.HasPrefix(id, "theme:"):
		return Action{Kind: ActTheme, Target: strings.TrimPrefix(id, "theme:")}
	case strings.HasPrefix(id, "lang:"):
		return Action{Kind: ActLang, Target: strings.TrimPrefix(id, "lang:")}
	case strings.HasPrefix(id, "report:"):
		parts := strings.SplitN(strings.TrimPrefix(id, "report:"), ":", 3)
		a := Action{Kind: ActReport}
		if len(parts) > 0 {
			a.Target = parts[0]
		}
		if len(parts) > 1 {
			a.Arg = parts[1]
		}
		if len(parts) > 2 {
			a.Extra = parts[2]
		}
		return a
	case strings.HasPrefix(id, "last:"):
		return Action{Kind: ActLast, Target: strings.TrimPrefix(id, "last:")}
	case strings.HasPrefix(id, "amt:"):
		return Action{Kind: ActBuy, Target: strings.TrimPrefix(id, "amt:")}
	case strings.HasPrefix(id, "bin:"):
		return Action{Kind: ActCode, Target: strings.TrimPrefix(id, "bin:")}
	case strings.HasPrefix(id, "lampLast:"):
		return Action{Kind: ActCmd, Target: "lampLast", Arg: strings.TrimPrefix(id, "lampLast:")}
	case strings.HasPrefix(id, "xbonStorno:"):
		return Action{Kind: ActCmd, Target: "storno", Arg: strings.TrimPrefix(id, "xbonStorno:")}
	case strings.HasPrefix(id, "cmd:"):
		rest := strings.TrimPrefix(id, "cmd:")
		name, arg, _ := strings.Cut(rest, ":")
		return Action{Kind: ActCmd, Target: name, Arg: arg}
	case id == "acPending":
		return Action{Kind: ActCmd, Target: "pending"}
	case id == "acLast":
		return Action{Kind: ActCmd, Target: "copy"}
	case id == "reportShow":
		return Action{Kind: ActCmd, Target: "reportShow"}
	case id == "reportPrint":
		return Action{Kind: ActCmd, Target: "reportPrint"}
	default:
		return Action{Kind: ActNav, Target: id}
	}
}
