// checksum.go provides SHA256 integrity verification for downloaded archives.
package installer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// ComputeChecksum returns the hex-encoded SHA256 digest of the file at filePath.
func ComputeChecksum(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("opening file for checksum: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("computing checksum: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// VerifyChecksum reads the file and compares its SHA256 hex digest to expected.
func VerifyChecksum(filePath, expected string) error {
	got, err := ComputeChecksum(filePath)
	if err != nil {
		return err
	}
	if got != expected {
		return fmt.Errorf("checksum mismatch: got %s, want %s", got, expected)
	}
	return nil
}
