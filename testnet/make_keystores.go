package main

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "strings"
    "time"

    "golang.org/x/crypto/scrypt"
    "golang.org/x/crypto/sha3"
)

type KeyFile struct {
    Address string                 `json:"address"`
    Crypto  map[string]interface{} `json:"crypto"`
    Id      string                 `json:"id"`
    Version int                    `json:"version"`
}

func encrypt(privBytes []byte, pass string) (map[string]interface{}, error) {
    salt := make([]byte, 32)
    if _, err := rand.Read(salt); err != nil {
        return nil, err
    }
    dk, err := scrypt.Key([]byte(pass), salt, 262144, 8, 1, 32)
    if err != nil {
        return nil, err
    }
    block, err := aes.NewCipher(dk[:16])
    if err != nil {
        return nil, err
    }
    iv := make([]byte, 16)
    if _, err := rand.Read(iv); err != nil {
        return nil, err
    }
    stream := cipher.NewCTR(block, iv)
    ciphertext := make([]byte, len(privBytes))
    stream.XORKeyStream(ciphertext, privBytes)

    // mac = keccak256(dk[16:32] || ciphertext)
    mach := sha3.NewLegacyKeccak256()
    mach.Write(dk[16:32])
    mach.Write(ciphertext)
    mac := hex.EncodeToString(mach.Sum(nil))

    cryptoMap := map[string]interface{}{
        "cipher": "aes-128-ctr",
        "cipherparams": map[string]string{"iv": hex.EncodeToString(iv)},
        "ciphertext": hex.EncodeToString(ciphertext),
        "kdf": "scrypt",
        "kdfparams": map[string]interface{}{
            "dklen": 32,
            "n":     262144,
            "r":     8,
            "p":     1,
            "salt":  hex.EncodeToString(salt),
        },
        "mac": mac,
    }
    return cryptoMap, nil
}

func main() {
    // Use the exact address <-> private key pairs you provided
    pairs := []struct{
        Address string
        PrivHex string
    }{
        {"0xBcA3d19C24a0ebFc02b4047977F9473C388e4E98", "d3d9f821e5df04816f0d97253c1d0a071de76c743d761e703fc50f638d6f5a27"},
        {"0x993669a7793F24b5F2e81c03dB494e0a83EAAE17", "b766e346d576615d4c476773be40f8fa98ebf458b0165f93aa57566bbab4ca03"},
        {"0x649A72A7c3b30a8a347dC7A549D3e50c3eD4c97c", "62f2b6e05996548b52966738e3485bed674e339e56a6e56cc55e2572f0371674"},
    }

    outdir := "validators/keystore"
    os.MkdirAll(outdir, 0700)

    for i, p := range pairs {
        priv, err := hex.DecodeString(p.PrivHex)
        if err != nil {
            log.Fatal(err)
        }
        // encrypt raw private key bytes with empty password
        cryptoMap, err := encrypt(priv, "")
        if err != nil {
            log.Fatal(err)
        }
        // normalize address: remove 0x and lowercase
        addr := p.Address
        if len(addr) >= 2 && addr[0:2] == "0x" {
            addr = addr[2:]
        }
        addr = strings.ToLower(addr)

        kf := KeyFile{
            Address: addr,
            Crypto:  cryptoMap,
            Id:      fmt.Sprintf("%d-%s", time.Now().UnixNano(), addr[:8]),
            Version: 3,
        }
        raw, _ := json.MarshalIndent(kf, "", "  ")
        name := fmt.Sprintf("UTC--%s--%s.json", time.Now().UTC().Format("2006-01-02T15-04-05.999999999Z"), addr)
        if err := ioutil.WriteFile(outdir+"/"+name, raw, 0600); err != nil {
            log.Fatal(err)
        }
        fmt.Println("Wrote keystore:", outdir+"/"+name)
        // create validator datadir structure
        vdir := fmt.Sprintf("validator%d", i+1)
        os.MkdirAll(vdir+"/keystore", 0700)
        ioutil.WriteFile(vdir+"/keystore/"+name, raw, 0600)
    }
    fmt.Println("Done: keystores in validators/keystore and validator1..3 datadirs")
}
