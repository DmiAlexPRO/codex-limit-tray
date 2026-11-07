package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codexlimittray/internal/update"
)

func TestSign(t *testing.T) {
	path := filepath.Join(t.TempDir(), update.AssetName)
	content := []byte("installer fixture")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	seed := base64.StdEncoding.EncodeToString(key.Seed())
	if err := sign(path, "1.2.3", seed, "wrong"); err == nil {
		t.Fatal("accepted mismatched key")
	}
	if err := sign(path, "1.2.3", seed, base64.StdEncoding.EncodeToString(pub)); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path + ".sig")
	if err != nil {
		t.Fatal(err)
	}
	sig, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(data)))
	msg := update.SignatureMessage("1.2.3", fmt.Sprintf("%x", sha256.Sum256(content)))
	if err != nil || !ed25519.Verify(pub, msg, sig) {
		t.Fatal("release signature does not verify")
	}
}
