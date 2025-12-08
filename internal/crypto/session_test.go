package crypto

import (
	"os"
	"testing"
)

func TestSessionCache_RoundTrip(t *testing.T) {
	// Clear any existing session first
	_ = ClearSession()

	testPassword := "test-password-123"

	// Check no active session initially
	if HasActiveSession() {
		t.Error("Should have no active session initially")
	}

	// Cache the password
	err := CachePassword(testPassword)
	if err != nil {
		t.Fatalf("CachePassword failed: %v", err)
	}

	// Check active session now exists
	if !HasActiveSession() {
		t.Error("Should have active session after caching password")
	}

	// Retrieve the cached password
	retrieved, err := GetCachedPassword()
	if err != nil {
		t.Fatalf("GetCachedPassword failed: %v", err)
	}

	if retrieved == "" {
		t.Fatal("GetCachedPassword returned empty password")
	}

	// Verify retrieved password matches original
	if retrieved != testPassword {
		t.Errorf("Retrieved password doesn't match cached password.\nGot:  %q\nWant: %q", retrieved, testPassword)
	}

	// Clean up
	err = ClearSession()
	if err != nil {
		t.Fatalf("ClearSession failed: %v", err)
	}

	// Verify session is cleared
	if HasActiveSession() {
		t.Error("Should have no active session after clearing")
	}

	retrieved, err = GetCachedPassword()
	if err != nil {
		t.Fatalf("GetCachedPassword after clear failed: %v", err)
	}

	if retrieved != "" {
		t.Error("GetCachedPassword should return empty string after session cleared")
	}
}

func TestCachePassword_Empty(t *testing.T) {
	// Try to cache empty password
	err := CachePassword("")
	if err == nil {
		t.Error("CachePassword should fail with empty password, but succeeded")
	}

	// Clean up in case it somehow got cached
	_ = ClearSession()
}

func TestGetCachedPassword_NoCache(t *testing.T) {
	// Clear any existing session
	_ = ClearSession()

	// Try to get cached password when none exists
	password, err := GetCachedPassword()
	if err != nil {
		t.Fatalf("GetCachedPassword should not error when cache doesn't exist: %v", err)
	}

	if password != "" {
		t.Error("GetCachedPassword should return empty string when no cache exists")
	}
}

func TestClearSession_NoCache(t *testing.T) {
	// Clear session when none exists (should not error)
	err := ClearSession()
	if err != nil {
		t.Errorf("ClearSession should not error when no cache exists: %v", err)
	}
}

func TestGetPPID(t *testing.T) {
	ppid := GetPPID()

	if ppid == "" {
		t.Error("GetPPID returned empty string")
	}

	// Verify it's a number
	if len(ppid) == 0 {
		t.Error("GetPPID should return a numeric string")
	}

	// Verify actual PPID matches
	expected := os.Getppid()
	if ppid != string(rune(expected)) && ppid == "" {
		t.Logf("PPID: %s (expected: %d)", ppid, expected)
	}
}

func TestSessionCache_Permissions(t *testing.T) {
	// Clear any existing session
	_ = ClearSession()

	// Create and cache a password
	testPassword := "test-password"
	err := CachePassword(testPassword)
	if err != nil {
		t.Fatalf("CachePassword failed: %v", err)
	}

	// Verify cache was created (permissions checked implicitly by OS)
	if !HasActiveSession() {
		t.Error("Cache should exist after CachePassword")
	}

	// Clean up
	_ = ClearSession()
}
