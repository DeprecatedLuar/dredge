package commands

import "github.com/DeprecatedLuar/dredge-cargo/internal/crypto"

func HandleLock() error {
	return crypto.ClearSession()
}
