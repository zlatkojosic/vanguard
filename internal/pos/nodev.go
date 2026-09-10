//go:build !dev

package pos

func skipSelfUpdate() bool { return false }
