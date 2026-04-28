//go:build windows

package config

func isOwnedByCurrentUser(path string) (bool, error) {
	return true, nil
}
