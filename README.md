# adssl-issuer

adssl-issuer is a kubernetes operator that provides a [cert-manager](https://cert-manager.io/) Issuer for Active Directory SSL certificates.
Certificates are requested from AD via the [adssl](https://github.com/tomdoherty/adssl) library

## Install

```shell
make docker-build docker-push IMG=adssl-operator:v0.1.0
make deploy IMG=adssl-operator:v0.1.0
```

## Usage

Create a secret for the Active Directory password

```shell
kubectl create secret generic adsslissuer-secret \
  --from-literal=password=t0p-s3cret
```

Configure AdsslIssuer

```yaml
---
apiVersion: certmanager.tomdoherty.io/v1alpha2
kind: AdsslIssuer
metadata:
  name: adsslissuer-sample
spec:
  url: adssl.example.com
  user: tom
  passwordRef:
    name: adsslissuer-secret
    key: password
```

Request a Certificate as normal

```yaml
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  name: kafka-tls
spec:
  # The secret name to store the signed certificate
  secretName: kafka-tls-secret
  # Common Name
  commonName: kafka
  # DNS SAN
  dnsNames:
    - kafka-0
  # IP Address SAN
  ipAddresses:
    - "127.0.0.1"
  # Duration of the certificate
  duration: 24h
  # Renew 8 hours before the certificate expiration
  renewBefore: 8h
  # The reference to the AdsslIssuer issuer
  issuerRef:
    group: certmanager.tomdoherty.io
    kind: CertificateRequest
    name: adsslissuer-sample
```
