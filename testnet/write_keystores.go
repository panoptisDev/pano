package main

import (
    "encoding/hex"
    "fmt"
    "io/ioutil"
    "log"
    "os"

    "github.com/ethereum/go-ethereum/accounts/keystore"
    "github.com/ethereum/go-ethereum/crypto"
)

func main() {
    keys := []string{
        "d3d9f821e5df04816f0d97253c1d0a071de76c743d761e703fc50f638d6f5a27",
        "b766e346d576615d4c476773be40f8fa98ebf458b0165f93aa57566bbab4ca03",
        "62f2b6e05996548b52966738e3485bed674e339e56a6e56cc55e2572f0371674",
    }

    outdir := "validators/keystore"
    os.MkdirAll(outdir, 0700)

    pass := "" // empty password for testing

    for i, k := range keys {
        priv, err := hex.DecodeString(k)
        if err != nil {
            log.Fatal(err)
        }
        pk, err := crypto.ToECDSA(priv)
        if err != nil {
            log.Fatal(err)
        }
        ks := keystore.NewKeyStore("./tmpks", keystore.StandardScryptN, keystore.StandardScryptP)
        account, err := ks.ImportECDSA(pk, pass)
        if err != nil {
            log.Fatal(err)
        }
        // read file created in tmpks
        files, _ := ioutil.ReadDir("./tmpks")
        for _, f := range files {
            data, _ := ioutil.ReadFile("./tmpks/" + f.Name())
            ioutil.WriteFile(fmt.Sprintf("%s/UTC--%d--%s", outdir, i+1, account.Address.Hex()[2:]), data, 0600)
            os.Remove("./tmpks/" + f.Name())
        }
    }
    os.RemoveAll("./tmpks")
    fmt.Println("Keystores written to validators/keystore")
}
