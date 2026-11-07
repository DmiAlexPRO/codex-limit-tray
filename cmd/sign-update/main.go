// Command sign-update is a maintainer tool; it is not shipped in the app.
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"codexlimittray/internal/update"
)

func main() {
	asset := flag.String("asset", "", "Windows release installer")
	version := flag.String("version", "", "release version without the v prefix")
	keyFile := flag.String("generate-key", "", "create a new private seed file offline and print only the public key")
	showPublic := flag.Bool("public-key", false, "derive the public key from UPDATE_SIGNING_SEED without printing the seed")
	flag.Parse()
	var err error
	switch {
	case *showPublic:
		var key ed25519.PrivateKey
		key, err = privateKey(os.Getenv("UPDATE_SIGNING_SEED"))
		if err == nil {
			fmt.Println(base64.StdEncoding.EncodeToString(key.Public().(ed25519.PublicKey)))
		}
	case *keyFile != "":
		err = generate(*keyFile)
	default:
		err = sign(*asset, *version, os.Getenv("UPDATE_SIGNING_SEED"), os.Getenv("UPDATE_PUBLIC_KEY"))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func generate(path string) error {
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(f, base64.StdEncoding.EncodeToString(key.Seed()))
	closeErr := f.Close()
	if err = errors.Join(err, closeErr); err != nil {
		_ = os.Remove(path)
		return err
	}
	fmt.Println(base64.StdEncoding.EncodeToString(pub))
	return nil
}

func sign(asset, version, seedText, publicText string) error {
	if filepath.Base(asset) != update.AssetName || version == "" || strings.ContainsAny(version, "\r\n") {
		return errors.New("expected the release installer and its exact version")
	}
	key, err := privateKey(seedText)
	if err != nil {
		return err
	}
	pub, err := base64.StdEncoding.DecodeString(strings.TrimSpace(publicText))
	if err != nil || subtle.ConstantTimeCompare(pub, key.Public().(ed25519.PublicKey)) != 1 {
		return errors.New("UPDATE_PUBLIC_KEY does not match signing seed")
	}
	f, err := os.Open(asset)
	if err != nil {
		return err
	}
	defer f.Close()
	hash := sha256.New()
	if _, err = io.Copy(hash, f); err != nil {
		return err
	}
	message := update.SignatureMessage(version, hex.EncodeToString(hash.Sum(nil)))
	sig := base64.StdEncoding.EncodeToString(ed25519.Sign(key, message))
	return os.WriteFile(asset+".sig", []byte(sig+"\n"), 0644)
}

func privateKey(seedText string) (ed25519.PrivateKey, error) {
	seed, err := base64.StdEncoding.DecodeString(strings.TrimSpace(seedText))
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil, errors.New("invalid UPDATE_SIGNING_SEED")
	}
	return ed25519.NewKeyFromSeed(seed), nil
}
