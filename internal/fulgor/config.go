package fulgor

import (
	"encoding/json"

	"nts/vanguard/internal/operators"
)

func ParseOperatorConfig(raw []byte) operators.Overlay {
	var body struct {
		OperatorConfig *struct {
			Currency  string `json:"currency"`
			Operators map[string]struct {
				Amounts []int `json:"amounts"`
			} `json:"operators"`
		} `json:"operatorConfig"`
	}
	if json.Unmarshal(raw, &body) != nil || body.OperatorConfig == nil {
		return operators.Overlay{}
	}
	ov := operators.Overlay{
		Currency: body.OperatorConfig.Currency,
		Amounts:  map[string][]int{},
	}
	for code, op := range body.OperatorConfig.Operators {
		if len(op.Amounts) == 0 {
			continue
		}
		ov.Amounts[code] = append([]int(nil), op.Amounts...)
	}
	return ov
}
