package cipherfns


import (
    "testing"
    "strings"
    "regexp"
)


var hexStrMatch = regexp.MustCompile(`^[0-9a-fA-F]+$`)


func checkErr(t *testing.T, err error, expectedSubStr string) {
    t.Helper()
    // unexpected error
    if expectedSubStr == "" && err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    // expected error but got nil
    if expectedSubStr != "" && err == nil{
        t.Fatalf("Expected error containing %q but got nil", expectedSubStr)
    }
    // substring dosen't match
    if expectedSubStr != "" && !strings.Contains(err.Error(), expectedSubStr){
        t.Errorf("Wrong error\nExpected:\t%q\nGot:\t\t%q", expectedSubStr, err)
    }
}


//{{{ Generate Random Hex
func TestGenerateRandomHex(t *testing.T) {
    tests := []struct {
        name            string
        size            int
        expectedSubStr  string
    }{
        {"Success1024",         1024,   ""},
        {"Success64",           64,     ""},
        {"Success32",           32,     ""},
        {"Success24",           24,     ""},
        {"Success16",           16,     ""},
        {"Success8",            8,      ""},
        {"FailNegativeSize",    -1,     "Invalid size"},
        {"FailOutOfBounds",     9999,   "Invalid size"},
    }
    // Iterate
    for _, tc := range tests{
        t.Run(tc.name, func(t *testing.T) {
            randHexStr, err := GenerateRandomHex(tc.size)
            // Check error
            checkErr(t, err, tc.expectedSubStr)
            // If err occurs no point in testing return value
            if err != nil {
                return
            }
            // Check if its hex str
            if !hexStrMatch.MatchString(randHexStr){
                t.Errorf("Output is not HEX string: %s", randHexStr)
            }
            // Check len
            if len(randHexStr) != tc.size * 2 {
                t.Errorf("Output is wrong size %q\nExpected:\t%d\nGot:\t\t%d", randHexStr, tc.size *2, len(randHexStr))
            }
        })
    }
}
//}}} Generate Random Hex


//{{{ Derivate Key
//{{{ python check for non belivers
/*
python3 -c "import hashlib, binascii; print(
    binascii.hexlify(
        hashlib.pbkdf2_hmac(
            'sha256',
            b'emotion_engine_xoxo',
            binascii.unhexlify('344feecf40d375380ed5f523b9029647bf7c9f2261e0341a87aa5df6d49c4e31'),
            100000,
            32
        )
    ).decode()
)"
*/
//}}} python check for non belivers
func TestDerivateKey(t *testing.T) {
    salt1 := "1fff99cff0d1751a09d5f521b80286f7bf7c8f2261901f1aa7aa5df6df8cf911"
    salt2 := "344feecf40d375380ed5f523b9029647bf7c9f2261e0341a87aa5df6d49c4e31"
    tests := []struct {
        name            string
        salt            string
        password        string
        expectedSubStr  string
        expectedKey     string
    }{
        {"Success", salt1, "cats_and_dogs123", "", "34703f2f8208765c2c2fa1590c6c1b6cfa83d777852248b0d2a1728e131ecf8a"},
        {"Success", salt2, "cats_and_dogs123", "", "5f1d3e25d1483b306f281dafccea5ba5f909046a2261a1f7809ecf22093d1b6b"},
        {"Success", salt1, "emotion_engine_xoxo", "", "40abe08e11aa7624315c0531cd85c7ec380e136c9716a15a143eaaf816cfeff2"},
        {"Success", salt2, "emotion_engine_xoxo", "", "25780cc3a2494b0a784f02a1eebad32bb06bdaadb34857668f54e0b566ca6da6"},
        {"FailSaltNotHexStr", "WrongSalt09xk:OK{", "emotion_engine_xoxo", "Failed to decode salt from hex", ""},
    }
    // Iterate
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            keyHexStr, err := DerivateKey(tc.salt, tc.password)
            // Check error
            checkErr(t, err, tc.expectedSubStr)
            // If err occurs no point in testing return value
            if err != nil {
                return
            }
            // Check if its hex str
            if !hexStrMatch.MatchString(keyHexStr){
                t.Errorf("Output is not HEX string: %s", keyHexStr)
            }
            // Check len
            if tc.expectedKey != keyHexStr{
                t.Errorf("Output is wrong:\nExpected:\t%q\nGot:\t\t%q", tc.expectedKey, keyHexStr)
            }

        })
    }
}
//}}} Derivate Key


//{{{ Encrypt Decrypt
func TestEncDecAESHex(t *testing.T) {
    tests := []struct {
        name            string
        keyHexStr       string
        plaintext       string
        expErrSubStr    string
    }{
        {
            name:           "Success",
            keyHexStr:      "1f3f99cff0d1951b09d5f521b8b284f7bf7c8f2d61901f1ac7aa5df6df8cf921",
            plaintext:      "Secret_MSG!",
            expErrSubStr:   "",
        }, {
            name:           "SuccessLongMsg",
            keyHexStr:      "1f3f99cff0d1951b09d5f521b8b284f7bf7c8f2d61901f1ac7aa5df6df8cf921",
            plaintext:      strings.Repeat("Secret_MSG!", 100),
            expErrSubStr:   "",
        }, {
            name:           "FailKeyNotHex",
            keyHexStr:      "not_hex_string",
            plaintext:      "Secret_MSG!",
            expErrSubStr:   "Failed to decode key from hex",
        },
    }
    // Iterate
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            cipherHexStr, err := EncryptAESHex(tc.keyHexStr, tc.plaintext)
            // Check error
            checkErr(t, err, tc.expErrSubStr)
            // If err occurs no point in testing return value
            if err != nil {
                return
            }
            plaintext, err := DecryptAESHex(tc.keyHexStr, cipherHexStr)
            // Check error
            checkErr(t, err, tc.expErrSubStr)
            // If err occurs no point in testing return value
            if err != nil {
                return
            }
            if tc.plaintext != plaintext {
                t.Errorf("Plaintext missmatch:\nInput:\t%q\nOutput:\t%q", tc.plaintext, plaintext)
            }
        })
    }
}
//}}} Encrypt Decrypt


//{{{ Encrypt/Decrypt Api Key
func TestEncDecApiKey(t *testing.T) {
    saltHexStr, err := GenerateRandomHex(32)
    if err != nil {
        t.Fatalf("Failed to generate salt: %v", err)
    }
    tests := []struct {
        name            string
        password        string
        apiKey          string
        expErrSubStr    string
    }{
        {
            name:           "Success",
            password:       "Raw_input_pwd!",
            apiKey:         "API_KEY_PRIVATE",
            expErrSubStr:   "",
        }, {
            name:           "SuccessLongPassword",
            password:       strings.Repeat("Raw_input_pwd!", 100),
            apiKey:         "API_KEY_PRIVATE",
            expErrSubStr:   "",
        },// No idea how can I make this one fail unless we inject some custom fn
    }
    // Iterate
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            encApiKeyHexStr, err := EncryptApiKey(saltHexStr, tc.password, tc.apiKey)
            // Check error
            checkErr(t, err, tc.expErrSubStr)
            // If err occurs no point in testing return value
            if err != nil {
                return
            }
            apiKey, err := DecryptApiKey(saltHexStr, tc.password, encApiKeyHexStr)
            // Check error
            checkErr(t, err, tc.expErrSubStr)
            // If err occurs no point in testing return value
            if err != nil {
                return
            }
            if apiKey != tc.apiKey {
                t.Errorf("apiKey missmatch:\nInput:\t%q\nOutput:\t%q", tc.apiKey, apiKey)
            }
        })
    }
}
//}}} Encrypt/Decrypt Api Key




