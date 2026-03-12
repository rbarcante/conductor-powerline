package segments

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rbarcante/conductor-powerline/internal/debug"
)

// workflowCacheEntry is the on-disk JSON structure for cached workflow data.
type workflowCacheEntry struct {
	Data     WorkflowData `json:"data"`
	StoredAt time.Time    `json:"stored_at"`
	TTL      string       `json:"ttl"`
}

// WorkflowFileCache persists WorkflowData to disk, keyed by workspace path hash.
// Writes are atomic (temp file + rename) so concurrent processes never read partial JSON.
type WorkflowFileCache struct {
	dir string
	ttl time.Duration
}

// NewWorkflowFileCache creates a file-based workflow cache rooted at dir with the given TTL.
func NewWorkflowFileCache(dir string, ttl time.Duration) *WorkflowFileCache {
	return &WorkflowFileCache{dir: dir, ttl: ttl}
}

// Store writes workflow data to a JSON file keyed by workspace identifier.
// Silently returns on any I/O error (graceful degradation).
func (wc *WorkflowFileCache) Store(key string, data *WorkflowData) {
	if err := os.MkdirAll(wc.dir, 0o700); err != nil {
		debug.Logf("workflow_cache", "cannot create cache dir: %v", err)
		return
	}

	entry := workflowCacheEntry{
		Data:     *data,
		StoredAt: time.Now(),
		TTL:      wc.ttl.String(),
	}

	b, err := json.Marshal(entry)
	if err != nil {
		debug.Logf("workflow_cache", "marshal error: %v", err)
		return
	}

	path := wc.keyPath(key)
	if err := wc.atomicWrite(path, b); err != nil {
		debug.Logf("workflow_cache", "write error: %v", err)
	}
}

// Get reads cached workflow data for the given key. Returns nil if the file
// does not exist or cannot be parsed. Sets IsStale if the TTL has expired.
func (wc *WorkflowFileCache) Get(key string) *WorkflowData {
	path := wc.keyPath(key)

	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var entry workflowCacheEntry
	if err := json.Unmarshal(b, &entry); err != nil {
		debug.Logf("workflow_cache", "unmarshal error for %s: %v", key, err)
		return nil
	}

	result := entry.Data
	if time.Since(entry.StoredAt) > wc.ttl {
		result.IsStale = true
	}
	return &result
}

// keyPath returns the file path for a given workspace key.
// The key is hashed with SHA-256 to produce a safe, fixed-length filename.
func (wc *WorkflowFileCache) keyPath(key string) string {
	h := sha256.Sum256([]byte(key))
	return filepath.Join(wc.dir, fmt.Sprintf("%x.workflow.json", h))
}

// atomicWrite writes data to a temp file then renames it to path, ensuring
// readers never see a partial write.
func (wc *WorkflowFileCache) atomicWrite(path string, data []byte) error {
	tmp, err := os.CreateTemp(wc.dir, ".tmp-wf-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}
