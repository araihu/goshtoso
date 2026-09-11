package iconpack

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"

	muambasource "github.com/araihu/muamba/source"
)

// fileData retains small generated documents or the identity of a private disk
// file. Every disk read verifies the bytes consumed, including publication.
type fileData struct {
	data []byte
	path string
	size int64
	hash string
}

func memoryFile(data []byte) fileData {
	return fileData{data: data, size: int64(len(data)), hash: hashBytes(data)}
}

func (f fileData) copy(ctx context.Context, w io.Writer) error {
	var r io.Reader = bytes.NewReader(f.data)
	if f.path != "" {
		file, err := os.Open(f.path)
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()
		info, err := file.Stat()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("source is not a regular file")
		}
		r = file
	}
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(w, h), cancelReader{ctx, io.LimitReader(r, f.size+1)})
	if err != nil {
		return err
	}
	if n != f.size || fmt.Sprintf("%x", h.Sum(nil)) != f.hash {
		return fmt.Errorf("source integrity mismatch")
	}
	return ctx.Err()
}

func (f fileData) read(ctx context.Context) ([]byte, error) {
	var b bytes.Buffer
	if err := f.copy(ctx, &b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (f fileData) decodeJSON(ctx context.Context, value any) error {
	data, err := f.read(ctx)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

type cancelReader struct {
	ctx context.Context
	r   io.Reader
}

func (r cancelReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}

func captureMuambaFiles(ctx context.Context, engine *muambasource.Engine, sources []resolvedConfigSource, staging, lockPath string) (map[string]fileData, []byte, error) {
	files := map[string]fileData{}
	var lockBytes []byte
	err := engine.Walk(ctx, nil, func(meta muambasource.File, r io.Reader) error {
		// Capture provenance while Walk still owns the namespace mutation lock.
		// Another cooperative Lock cannot replace the lockfile between the
		// source snapshot and this read.
		if lockBytes == nil {
			var err error
			lockBytes, err = os.ReadFile(lockPath)
			if err != nil {
				return fmt.Errorf("read iconpack lock: %w", err)
			}
		}
		key, err := snapshotKey(meta.Path, sources)
		if err != nil {
			return err
		}
		if _, exists := files[key]; exists {
			return fmt.Errorf("iconpack source files resolve to duplicate path %q", key)
		}
		file, err := os.CreateTemp(staging, "source-*")
		if err != nil {
			return err
		}
		h := sha256.New()
		n, copyErr := io.Copy(io.MultiWriter(file, h), r)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		if n != meta.Size {
			return fmt.Errorf("source size mismatch")
		}
		files[key] = fileData{path: file.Name(), size: n, hash: fmt.Sprintf("%x", h.Sum(nil))}
		return nil
	})
	return files, lockBytes, err
}
