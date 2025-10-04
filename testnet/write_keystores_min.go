package main

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/ecdsa"
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "time"

    "golang.org/x/crypto/scrypt"
    "github.com/ethereum/go-ethereum/crypto"
)

type KeyFile struct {
    Address string                 `json:"address"`
    Crypto  map[string]interface{} `json:"crypto"`
    Id      string                 `json:"id"`
    Version int                    `json:"version"`
}

func encryptPrivateKey(priv *ecdsa.PrivateKey, pass string) ([]byte, map[string]interface{}, error) {
    privBytes := crypto.FromECDSA(priv)
    salt := make([]byte, 32)
    rand.Read(salt)
    dk, err := scrypt.Key([]byte(pass), salt, 262144, 8, 1, 32)
    if err != nil {
        return nil, nil, err
    }
    block, err := aes.NewCipher(dk)
    if err != nil {
        return nil, nil, err
    }
    iv := make([]byte, aes.BlockSize)
    rand.Read(iv)
    stream := cipher.NewCTR(block, iv)
    cipherText := make([]byte, len(privBytes))
    stream.XORKeyStream(cipherText, privBytes)

    cryptoMap := map[string]interface{}{
        "cipher": "aes-128-ctr",
        "cipherparams": map[string]string{"iv": hex.EncodeToString(iv)},
        "ciphertext": hex.EncodeToString(cipherText),
        "kdf": "scrypt",
        "kdfparams": map[string]interface{}{
            "dklen": 32,
            "n":     262144,
            "r":     8,
            "p":     1,
            "salt":  hex.EncodeToString(salt),
        },
        "mac": "",
    }
    return privBytes, cryptoMap, nil
}

func main() {
    keys := []string{
        "d3d9f821e5df04816f0d97253c1d0a071de76c743d761e703fc50f638d6f5a27",
        "b766e346d576615d4c476773be40f8fa98ebf458b0165f93aa57566bbab4ca03",
        "62f2b6e05996548b52966738e3485bed674e339e56a6e56cc55e2572f0371674",
    }
    outdir := "validators/keystore"
    os.MkdirAll(outdir, 0700)

    for i, k := range keys {
        priv, err := hex.DecodeString(k)
        if err != nil {
            log.Fatal(err)
        }
        pk, err := crypto.ToECDSA(priv)
        if err != nil {
            log.Fatal(err)
        }
        addr := crypto.PubkeyToAddress(pk.PublicKey).Hex()[2:]
        _, cryptoMap, err := encryptPrivateKey(pk, "")
        if err != nil {
            log.Fatal(err)
        }
        kf := KeyFile{
            Address: addr,
            Crypto:  cryptoMap,
            Id:      fmt.Sprintf("%d-%s", time.Now().UnixNano(), addr[:8]),
            Version: 3,
        }
        raw, _ := json.MarshalIndent(kf, "", "  ")
        name := fmt.Sprintf("UTC--%s--%s.json", time.Now().UTC().Format("2006-01-02T15-04-05.999999999Z"), addr)
        ioutil.WriteFile(outdir+"/"+name, raw, 0600)
        fmt.Println("Wrote keystore:", outdir+"/"+name)
        time.Sleep(100 * time.Millisecond)
    }

    fmt.Println("Done")
}
