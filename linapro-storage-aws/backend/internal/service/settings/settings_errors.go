package settings

import "lina-core/pkg/bizerr"

// ErrUnavailable returns the stable unavailable code for factory use.
func ErrUnavailable() error {
	return bizerr.NewCode(CodeStorageUnavailable)
}
