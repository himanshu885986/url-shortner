package encoding

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func Base62Encode(num uint64) string {
	if num == 0 {
		return string(alphabet[0])
	}
	var encoded strings.Builder
	for num > 0 {
		rem := num % 62
		encoded.WriteByte(alphabet[rem])
		num /= 62
	}
	// reverse
	s := encoded.String()
	r := []byte(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// MD5Hex returns the MD5 hash of s in hex (not recommended for security uses).
func MD5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

// SHA256Hex returns the SHA-256 hash of s in hex.
func SHA256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
