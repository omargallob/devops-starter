package installer

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// Extract extracts archivePath into destDir using the given format.
// stripComponents removes leading path elements from extracted entries.
func Extract(archivePath string, destDir string, format tooldef.ArchiveFormat, stripComponents int) error {
	switch format {
	case tooldef.FormatTarGz:
		return extractTarGz(archivePath, destDir, stripComponents)
	case tooldef.FormatTarXz:
		return extractTarXz(archivePath, destDir, stripComponents)
	case tooldef.FormatZip:
		return extractZip(archivePath, destDir, stripComponents)
	case tooldef.FormatBinary:
		return copyBinary(archivePath, destDir)
	default:
		return fmt.Errorf("unsupported archive format: %s", format)
	}
}

func extractTarGz(archivePath, destDir string, strip int) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gr.Close()

	return extractTar(tar.NewReader(gr), destDir, strip)
}

func extractTarXz(archivePath, destDir string, strip int) error {
	cmd := exec.Command("xz", "-d", "--stdout", archivePath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting xz: %w", err)
	}

	err = extractTar(tar.NewReader(stdout), destDir, strip)
	if err2 := cmd.Wait(); err2 != nil && err == nil {
		err = fmt.Errorf("xz process: %w", err2)
	}
	return err
}

func extractTar(tr *tar.Reader, destDir string, strip int) error {
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		name := stripPath(hdr.Name, strip)
		if name == "" {
			continue
		}

		target := filepath.Join(destDir, name)
		// Prevent path traversal
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			continue
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
		}
	}
}

func extractZip(archivePath, destDir string, strip int) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		name := stripPath(f.Name, strip)
		if name == "" {
			continue
		}

		target := filepath.Join(destDir, name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0o755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func copyBinary(archivePath, destDir string) error {
	data, err := os.ReadFile(archivePath)
	if err != nil {
		return err
	}
	// Use the base name of the archive as the binary name
	name := filepath.Base(archivePath)
	return os.WriteFile(filepath.Join(destDir, name), data, 0o755)
}

// stripPath removes the first n path components from a file path.
func stripPath(name string, n int) string {
	if n <= 0 {
		return name
	}
	parts := strings.SplitN(filepath.ToSlash(name), "/", n+1)
	if len(parts) <= n {
		return ""
	}
	return parts[n]
}
