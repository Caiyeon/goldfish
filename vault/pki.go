package vault

import (
    "crypto/tls"
    "errors"
)

func FetchCertificate(path, url string) (*tls.Certificate, error) {
    // initialize a client with goldfish's token
    client, err := NewGoldfishVaultClient()
    if err != nil {
        return nil, err
    }

    // write to pki role path
    resp, err := client.Logical().Write(path,
        map[string]interface{}{
            "common_name": url,
            "format":      "pem",
        })
    if err != nil {
        return nil, err
    }
    if resp.Data == nil {
        return nil, errors.New("No certificate issued.")
    }

    // parse body for certificate
    certRaw, ok := resp.Data["certificate"]
    if !ok {
        return nil, errors.New("Certificate not found in response")
    }
    keyRaw, ok := resp.Data["private_key"]
    if !ok {
        return nil, errors.New("Private key not found in response")
    }

    cert, ok := certRaw.(string)
    key, ok := keyRaw.(string)
    if cert == "" || key == "" {
        return nil, errors.New("Cert and key could not be asserted to string")
    }

    pair, err := tls.X509KeyPair([]byte(cert), []byte(key))
    if err != nil {
        return nil, err
    }

    return &pair, nil
}
