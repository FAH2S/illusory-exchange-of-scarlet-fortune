package cipherfns


import (
    "fmt"
    "crypto/aes"
    "crypto/cipher"
    "crypto/sha256"
    "crypto/rand"
    "encoding/hex"
)
import (
    "golang.org/x/crypto/pbkdf2"
)

//{{{ Generate Random Hex
// Generate random bytes and convert to hex string size = 32
func GenerateRandomHex(bytesSize int) (string, error) {
    if (bytesSize <= 0 || bytesSize > 1024) {
        return "", fmt.Errorf("GenerateRandomHex: Invalid size: must be in range [1, 1024]")
    }
    bytes := make([]byte, bytesSize)
    if _, err := rand.Read(bytes); err != nil {
        return "", fmt.Errorf("Failed to generate randomo bytes: %v", err)
    }
    return hex.EncodeToString(bytes), nil
}
//}}} Generate Random Hex


//{{{ Derivate Key
// Derivate key from salt(hex string) and 'password'
func DerivateKey(saltHexStr, password string) (string, error) {
    iterations := 100_000
    saltBytes, err := hex.DecodeString(saltHexStr)
    if err != nil {
        return "", fmt.Errorf("Failed to decode salt from hex: %v", err)
    }

    // 32 bytes = 64 hexchar = 256 bits, for AES-256
    keyLen := 32
    keyBytes := pbkdf2.Key([]byte(password), saltBytes, iterations, keyLen, sha256.New)
    // Convert bytes to hex string
    keyHexStr := hex.EncodeToString(keyBytes)

    return keyHexStr, nil
}
//}}} Derivate Key


//{{{ Encrypt
// Encrypt, ex.: enc(key, apiKey) = encApiKey
func EncryptAESHex(keyHexStr, plaintext string) (string, error) {
    // Convert hex string to bytes
    keyBytes, err := hex.DecodeString(keyHexStr)
    if err != nil {
        return "", fmt.Errorf("Failed to decode key from hex: %v", err)
    }

    cipher, err := EncryptAES(keyBytes, []byte(plaintext))
    if err != nil {
        return "", fmt.Errorf("Failed to encrypt: %v", err)
    }

    // Return hex string of cipher
    return hex.EncodeToString(cipher), nil
}

func EncryptAES(keyBytes, plaintextBytes []byte) ([]byte, error) {
    // AES cipher block
    block, err := aes.NewCipher(keyBytes)
    if err != nil {
        return nil, fmt.Errorf("Failed to create cipher block: %v", err)
    }

    // GCM mode
    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return nil, fmt.Errorf("Failed to create GCM mode: %v", err)
    }

    // Nonce
    nonce := make([]byte, aesGCM.NonceSize())
    rand.Read(nonce)

    // Encrypt (plaintext raw string)
    ciphertext := aesGCM.Seal(nil, nonce, plaintextBytes, nil)
    // Final = nonce + ciphertext
    final := append(nonce, ciphertext...)
    return final, nil
}
//}}} Encrypt


//{{{ Decrypt
func DecryptAESHex(keyHexStr, cipherHexStr string) (string, error) {
    // Conver hex tring to bytes
    keyBytes, err := hex.DecodeString(keyHexStr)
    if err != nil {
        return "", fmt.Errorf("Failed to decode key from hex: %v", err)
    }
    cipherBytes, err := hex.DecodeString(cipherHexStr)
    if err != nil {
        return "", fmt.Errorf("Failed to decode cipher from hex: %v", err)
    }

    plaintextBytes, err := DecryptAES(keyBytes, cipherBytes)
    if err != nil {
        return "", fmt.Errorf("Failed to decrypt: %v", err)
    }

    // Return plain plaintext (no bytes, no hex string)
    return string(plaintextBytes), nil
}


func DecryptAES(keyBytes, cipherBytes []byte) ([]byte, error) {
    // AES cipher block
    block, err := aes.NewCipher(keyBytes)
    if err != nil {
        return nil, fmt.Errorf("Failed to create cipher block: %v", err)
    }

    // GCM mode
    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return nil, fmt.Errorf("Failed to create GCM mode: %v", err)
    }

    nonceSize := aesGCM.NonceSize()
    if len(cipherBytes) < nonceSize {
        return nil, fmt.Errorf("Invalid ciphertext: length is smaller than nonce size")
    }
    // cipher = nonce + actual ciphertext
    nonce := cipherBytes[:nonceSize]
    ciphertextBytes := cipherBytes[nonceSize:]

    decryptedBytes, err := aesGCM.Open(nil, nonce, ciphertextBytes, nil)
    if err != nil {
        return nil, fmt.Errorf("Failed to decrypt: %v", err)
    }
    return decryptedBytes, nil
}
//}}} Decrypt


//{{{ Encrypt/Decrypt Api Key
func EncryptApiKey(password, apiKey string) (string, string, error) {
    saltHexStr, err := GenerateRandomHex(32)
    if err != nil {
        return "", "", fmt.Errorf("Failed to generate salt: %v", err)
    }

    keyHexStr, err := DerivateKey(saltHexStr, password)
    if err != nil {
        return "", "", fmt.Errorf("Failed to derivate key: %v", err)
    }

    encApiKeyHexStr, err := EncryptAESHex(keyHexStr, apiKey)
    if err != nil {
        return "", "", fmt.Errorf("Failed to encrypt: %v", err)
    }

    return saltHexStr, encApiKeyHexStr, nil
}


func DecryptApiKey(saltHexStr, password, encApiKeyHexStr string) (string, error) {
    keyHexStr, err := DerivateKey(saltHexStr, password)
    if err != nil {
        return "", fmt.Errorf("Failed to derivate key: %v", err)
    }

    apiKey, err := DecryptAESHex(keyHexStr, encApiKeyHexStr)
    if err != nil {
        return "", fmt.Errorf("Failed to decrypt: %v", err)
    }

    return apiKey, nil
}
//}}} Encrypt/Decrypt Api Key


