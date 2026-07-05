package pet

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// SelfUpdate downloads the given release tag and replaces the current binaries.
func SelfUpdate(tag string) error {
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	url := fmt.Sprintf(
		"https://github.com/jansuthacheeva/ccpetline/releases/download/%s/ccpetline-%s-%s.%s",
		tag, runtime.GOOS, runtime.GOARCH, ext,
	)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("downloading: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading download: %w", err)
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding executable: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("resolving symlinks: %w", err)
	}
	installDir := filepath.Dir(exe)

	// Extract to a temp dir first so a corrupt archive never leaves the
	// user without binaries. Same filesystem as installDir so the final
	// renames cannot fail with a cross-device error.
	tmpDir, err := os.MkdirTemp(installDir, ".ccpetline-update-")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if runtime.GOOS == "windows" {
		err = extractZip(data, tmpDir)
	} else {
		err = extractTarGz(data, tmpDir)
	}
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("reading extracted files: %w", err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("archive contained no files")
	}
	for _, e := range entries {
		dest := filepath.Join(installDir, e.Name())
		// Windows cannot rename over an existing file.
		_ = os.Remove(dest)
		if err := os.Rename(filepath.Join(tmpDir, e.Name()), dest); err != nil {
			return fmt.Errorf("installing %s: %w", e.Name(), err)
		}
	}
	return nil
}

func extractTarGz(data []byte, dir string) error {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		name := filepath.Base(hdr.Name)
		dest := filepath.Join(dir, name)
		f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode)|0o755)
		if err != nil {
			return fmt.Errorf("creating %s: %w", name, err)
		}
		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return fmt.Errorf("writing %s: %w", name, err)
		}
		f.Close()
	}
	return nil
}

func extractZip(data []byte, dir string) error {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("zip: %w", err)
	}
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := filepath.Base(f.Name)
		dest := filepath.Join(dir, name)
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("opening %s: %w", name, err)
		}
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode()|0o755)
		if err != nil {
			rc.Close()
			return fmt.Errorf("creating %s: %w", name, err)
		}
		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			return fmt.Errorf("writing %s: %w", name, err)
		}
		out.Close()
		rc.Close()
	}
	return nil
}
