package oauth

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadRotatedToken_NoFile(t *testing.T) {
	dir := t.TempDir()
	creds, err := LoadRotatedToken(dir)
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %v", err)
	}
	if creds != nil {
		t.Errorf("expected nil credentials, got: %+v", creds)
	}
}

func TestLoadRotatedToken_ValidFile(t *testing.T) {
	dir := t.TempDir()
	content := `{"access_token":"stored-access","refresh_token":"stored-refresh","rotated_at":"` +
		time.Now().Format(time.RFC3339) + `"}`
	if err := os.WriteFile(filepath.Join(dir, "rotated-token.json"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	creds, err := LoadRotatedToken(dir)
	if err != nil {
		t.Fatalf("expected credentials, got error: %v", err)
	}
	if creds.AccessToken != "stored-access" {
		t.Errorf("expected stored-access, got %q", creds.AccessToken)
	}
	if creds.RefreshToken != "stored-refresh" {
		t.Errorf("expected stored-refresh, got %q", creds.RefreshToken)
	}
}

func TestLoadRotatedToken_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "rotated-token.json"), []byte("{bad json"), 0600); err != nil {
		t.Fatal(err)
	}

	creds, err := LoadRotatedToken(dir)
	if err != nil {
		t.Fatalf("expected nil error for corrupt file (graceful), got: %v", err)
	}
	if creds != nil {
		t.Errorf("expected nil credentials for corrupt file, got: %+v", creds)
	}
}

func TestLoadRotatedToken_ExpiredFile(t *testing.T) {
	dir := t.TempDir()
	expired := time.Now().Add(-8 * 24 * time.Hour) // 8 days ago
	content := `{"access_token":"old-access","refresh_token":"old-refresh","rotated_at":"` +
		expired.Format(time.RFC3339) + `"}`
	if err := os.WriteFile(filepath.Join(dir, "rotated-token.json"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	creds, err := LoadRotatedToken(dir)
	if err != nil {
		t.Fatalf("expected nil error for expired file, got: %v", err)
	}
	if creds != nil {
		t.Errorf("expected nil credentials for expired file, got: %+v", creds)
	}
}

func TestStoreRotatedToken(t *testing.T) {
	dir := t.TempDir()
	creds := &TokenCredentials{
		AccessToken:  "new-access",
		RefreshToken: "new-refresh",
	}

	if err := StoreRotatedToken(dir, creds); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	// Verify file exists and has correct permissions
	path := filepath.Join(dir, "rotated-token.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected permissions 0600, got %o", info.Mode().Perm())
	}

	// Verify contents can be loaded back
	loaded, err := LoadRotatedToken(dir)
	if err != nil {
		t.Fatalf("expected to load stored token: %v", err)
	}
	if loaded.AccessToken != "new-access" {
		t.Errorf("expected new-access, got %q", loaded.AccessToken)
	}
	if loaded.RefreshToken != "new-refresh" {
		t.Errorf("expected new-refresh, got %q", loaded.RefreshToken)
	}
}

func TestTryRotationLock_AcquireAndRelease(t *testing.T) {
	dir := t.TempDir()

	acquired, release := TryRotationLock(dir)
	if !acquired {
		t.Fatal("expected to acquire lock")
	}

	// Second attempt should fail
	acquired2, _ := TryRotationLock(dir)
	if acquired2 {
		t.Fatal("expected second lock attempt to fail")
	}

	// Release and try again
	release()
	acquired3, release3 := TryRotationLock(dir)
	if !acquired3 {
		t.Fatal("expected to acquire lock after release")
	}
	release3()
}

func TestTryRotationLock_StaleLockCleanup(t *testing.T) {
	dir := t.TempDir()

	// Create a stale lock file
	lockPath := filepath.Join(dir, "rotated-token.json.lock")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Backdate it to appear stale
	staleTime := time.Now().Add(-(staleLockAge + 1*time.Second))
	if err := os.Chtimes(lockPath, staleTime, staleTime); err != nil {
		t.Fatal(err)
	}

	// Should succeed because the lock is stale
	acquired, release := TryRotationLock(dir)
	if !acquired {
		t.Fatal("expected to acquire lock after stale lock cleanup")
	}
	release()
}
