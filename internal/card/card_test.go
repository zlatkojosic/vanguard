package card

import "testing"

func TestLuhnExample(t *testing.T) {
	if !ValidateLuhn("000123456600") {
		t.Fatal("expected valid")
	}
}

func TestChecksumAppend(t *testing.T) {
	card := "000123456"
	sum := CalculateCheckSum(card)
	got := card + itoa(sum) + "00"
	if got != "000123456600" {
		t.Fatalf("got %s", got)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [4]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
