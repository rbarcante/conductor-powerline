package oauth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/rbarcante/conductor-powerline/internal/debug"
)

const rotatedTokenFile = "rotated-token.json"
const rotatedTokenLockFile = "rotated-token.json.lock"

// rotatedTokenMaxAge is the maximum age for a rotated token before it's
// considered expired. Claude Code may have rotated its own tokens since then.
const rotatedTokenMaxAge = 7 * 24 * time.Hour

// rotatedTokenDir is the cache directory for rotated tokens.
// Set via SetRotatedTokenDir before calling GetCredentials.
var rotatedTokenDir string

// SetRotatedTokenDir sets the cache directory where rotated tokens are stored.
func SetRotatedTokenDir(dir string) {
	rotatedTokenDir = dir
}

// rotatedTokenEntry is the on-disk JSON structure for a rotated token.
type rotatedTokenEntry struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	RotatedAt    time.Time `json:"rotated_at"`
}

// LoadRotatedToken reads a previously rotated token from disk.
// Returns nil, nil if the file doesn't exist, is corrupt, or is expired.
func LoadRotatedToken(cacheDir string) (*TokenCredentials, error) {
	path := filepath.Join(cacheDir, rotatedTokenFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, nil
	}

	var entry rotatedTokenEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		debug.Logf("rotatedtoken", "corrupt rotated token file: %v", err)
		return nil, nil
	}

	if time.Since(entry.RotatedAt) > rotatedTokenMaxAge {
		debug.Logf("rotatedtoken", "rotated token expired (age %v)", time.Since(entry.RotatedAt))
		return nil, nil
	}

	return &TokenCredentials{
		AccessToken:  entry.AccessToken,
		RefreshToken: entry.RefreshToken,
	}, nil
}

// StoreRotatedToken writes a rotated token to disk atomically with 0600 permissions.
func StoreRotatedToken(cacheDir string, creds *TokenCredentials) error {
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		return err
	}

	entry := rotatedTokenEntry{
		AccessToken:  creds.AccessToken,
		RefreshToken: creds.RefreshToken,
		RotatedAt:    time.Now(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	path := filepath.Join(cacheDir, rotatedTokenFile)

	// Atomic write: temp file + rename
	tmp, err := os.CreateTemp(cacheDir, ".tmp-rotated-*")
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

// TryRotationLock attempts to acquire an exclusive lock for token rotation.
// Returns (true, releaseFn) on success; (false, nil) if already locked.
// Stale locks older than staleLockAge are automatically removed.
func TryRotationLock(cacheDir string) (bool, func()) {
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		return false, nil
	}

	lockPath := filepath.Join(cacheDir, rotatedTokenLockFile)
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		// Lock exists — check if stale
		info, statErr := os.Stat(lockPath)
		if statErr == nil && time.Since(info.ModTime()) > staleLockAge {
			debug.Logf("rotatedtoken", "removing stale rotation lock (age %v)", time.Since(info.ModTime()))
			os.Remove(lockPath)
			// Retry once
			f, err = os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL, 0o600)
			if err != nil {
				return false, nil
			}
		} else {
			return false, nil
		}
	}
	f.Close()
	release := func() {
		os.Remove(lockPath)
	}
	return true, release
}
