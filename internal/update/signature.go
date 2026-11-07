package update

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"strings"
)

const AssetName = "CodexLimitTray-windows-x64-Setup.exe"

// SignatureMessage binds a digest to the exact application, version and asset.
// version is the canonical tag without its leading v; digest is lowercase hex.
func SignatureMessage(version, digest string) []byte {
	return []byte("CodexLimitTray update v1\n" + version + "\n" + AssetName + "\n" + digest + "\n")
}

func decodeKey(encoded string) (ed25519.PublicKey, error) {
	key, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if err != nil || len(key) != ed25519.PublicKeySize {
		return nil, errors.New("доверенный ключ обновлений отсутствует или некорректен")
	}
	return ed25519.PublicKey(key), nil
}

func decodeSignature(data []byte) ([]byte, error) {
	s, ok := strings.CutSuffix(string(data), "\r\n")
	if !ok {
		s = strings.TrimSuffix(s, "\n")
	}
	if len(s) != base64.StdEncoding.EncodedLen(ed25519.SignatureSize) {
		return nil, errors.New("некорректная подпись установщика")
	}
	sig, err := base64.StdEncoding.Strict().DecodeString(s)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return nil, errors.New("некорректная подпись установщика")
	}
	return sig, nil
}
