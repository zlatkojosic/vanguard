package operators

import "strings"

type binbinOp struct{}

func (binbinOp) Owns(id string) bool {
	return strings.HasPrefix(id, "binbin")
}

func (binbinOp) View(id string, t Text) Screen {
	return flowScreen(id, t)
}

func (binbinOp) Submit(id string, form map[string]string, rt Runtime) {
	if strings.HasSuffix(id, ".charge") {
		runCharge(id, form["stid"], rt)
	}
}

func (binbinOp) Tap(id string, act Action, rt Runtime) bool {
	switch act.Kind {
	case ActLast:
		runLast(act.Target, rt)
		return true
	case ActCode:
		runCodeBuy(id, act.Target, rt)
		return true
	default:
		return false
	}
}

func binbinProducts() Screen {
	s := Screen{Kind: "menu", Columns: 2, Items: []Item{
		codeItem("100019", "BinBin 10,00 KM\n10 KM"),
		codeItem("100042", "Bin Pass -\n dnevni (150 min)\n16.50 KM"),
		codeItem("100044", "Bin Pass -\n mjesečni (400 min)\n75 KM"),
	}}
	return s
}
