package vault

import (
    "crypto/tls"
    "errors"
)

func FetchCertificate(path string, body map[string]interface{}) (*tls.Certificate, error) {
    // initialize a client with goldfish's token
    client, err := NewGoldfishVaultClient()
    if err != nil {
        return nil, err
    }

    // write to pki role path
    resp, err := client.Logical().Write(path, body)
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
    issuingCACertRaw, ok := resp.Data["issuing_ca"]
    if !ok {
        return nil, errors.New("Issuing CA Certificate not found in response")
    }
    keyRaw, ok := resp.Data["private_key"]
    if !ok {
        return nil, errors.New("Private key not found in response")
    }

    cert, ok := certRaw.(string)
    issuing_ca_cert, ok := issuingCACertRaw.(string)
    key, ok := keyRaw.(string)
    if cert == "" || key == "" || issuing_ca_cert == "" {
        return nil, errors.New("Cert, Issuing CA Cert, and Key could not be asserted to string")
    }

    pair, err := tls.X509KeyPair([]byte(cert + "\n" + issuing_ca_cert), []byte(key))
    if err != nil {
        return nil, err
    }

    return &pair, nil
}
