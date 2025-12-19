package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyGen(t *testing.T) {
	t.Run("generates key of correct size", func(t *testing.T) {
		key, err := KeyGen()
		require.NoError(t, err)
		assert.Equal(t, KeySize, len(key))
	})
}

func TestEncrypt(t *testing.T) {
	t.Run("successfully encrypts data", func(t *testing.T) {
		key, err := KeyGen()
		require.NoError(t, err)

		plaintext := []byte("test data")
		ciphertext, err := Encrypt(key, plaintext)
		require.NoError(t, err)
		assert.NotEmpty(t, ciphertext)
	})

	t.Run("encrypts same plaintext differently each time", func(t *testing.T) {
		key, err := KeyGen()
		require.NoError(t, err)

		plaintext := []byte("test data")
		ciphertext1, err := Encrypt(key, plaintext)
		require.NoError(t, err)
		ciphertext2, err := Encrypt(key, plaintext)
		require.NoError(t, err)

		assert.NotEqual(t, ciphertext1, ciphertext2)
	})

	t.Run("fails with invalid key size", func(t *testing.T) {
		invalidKey := []byte("short")
		plaintext := []byte("test data")

		_, err := Encrypt(invalidKey, plaintext)
		assert.Error(t, err)
	})

	t.Run("encrypts empty data", func(t *testing.T) {
		key, err := KeyGen()
		require.NoError(t, err)

		plaintext := []byte("")
		ciphertext, err := Encrypt(key, plaintext)
		require.NoError(t, err)
		assert.NotEmpty(t, ciphertext)
	})
}

func TestDecrypt(t *testing.T) {
	t.Run("successfully decrypts data", func(t *testing.T) {
		key, err := KeyGen()
		require.NoError(t, err)

		plaintext := []byte("test data")
		ciphertext, err := Encrypt(key, plaintext)
		require.NoError(t, err)

		decrypted, err := Decrypt(key, ciphertext)
		require.NoError(t, err)

		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("fails with wrong key", func(t *testing.T) {
		key1, err := KeyGen()
		require.NoError(t, err)
		key2, err := KeyGen()
		require.NoError(t, err)

		plaintext := []byte("test data")
		ciphertext, err := Encrypt(key1, plaintext)
		require.NoError(t, err)

		_, err = Decrypt(key2, ciphertext)
		assert.Error(t, err)
	})

	t.Run("fails with invalid key size", func(t *testing.T) {
		invalidKey := []byte("short")
		ciphertext := []byte("fake ciphertext")

		_, err := Decrypt(invalidKey, ciphertext)
		assert.Error(t, err)
	})
}
