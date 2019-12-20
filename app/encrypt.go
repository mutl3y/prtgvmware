/*
 * Copyright © 2019.  mutl3y
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package app

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
)

func createHash(key string) []byte {
	// We use an md5 hash here because aes requires a 32 byte key
	// It doesn't really matter that md5 is not secure because we're not storing the output
	hash := md5.Sum([]byte(key))
	dst := make([]byte, hex.EncodedLen(len(hash)))
	hex.Encode(dst, hash[:])
	return dst
}

// Encrypt encrypts data using the passphrase.
func Encrypt(data []byte, passphrase string) ([]byte, error) {
	block, err := aes.NewCipher(createHash(passphrase))
	if err != nil {
		return []byte{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return []byte{}, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return []byte{}, err
	}
	ciphered := gcm.Seal(nonce, nonce, data, nil)
	return ciphered, nil
}

// Decrypt decrypts data using the passphrase.
func Decrypt(data []byte, passphrase string) ([]byte, error) {
	key := createHash(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return []byte{}, err
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphered := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphered, nil)
	if err != nil {
		return []byte{}, err
	}
	return plaintext, nil
}
