# aserto-idp-sync

Aserto IDP Sync


config.yaml

```
---
logging:
  prod: true
  log_level: info

api:
  grpc:
    listen_address: "0.0.0.0:8282"
    connection_timeout_seconds: 10
    certs:
      tls_cert_path: "/home/root/.config/aserto/aserto-idp-sync/certs/grpc.crt"
      tls_key_path: "/home/root/.config/aserto/aserto-idp-sync/certs/grpc.key"
      tls_ca_cert_path: "/home/root/.config/aserto/aserto-idp-sync/certs/grpc-ca.crt"
  gateway:
    listen_address: "0.0.0.0:8383"
    allowed_origins:
    - https://0.0.0.0",
    - https://0.0.0.0:*"
    certs:
      tls_cert_path: "/home/root/.config/aserto/aserto-idp-sync/certs/gateway.crt"
      tls_key_path: "/home/root/.config/aserto/aserto-idp-sync/certs/gateway.key"
      tls_ca_cert_path: "/home/root/.config/aserto/aserto-idp-sync/certs/gateway-ca.crt"
  health:
    listen_address: "0.0.0.0:8484"

idp:
  auth0:
    domain: "<<<redacted>>>.us.auth0.com"
    client_id: "<<<redacted>>>"
    client_secret: "<<<redacted>>>"

directory:
  host_address: "authorizer.prod.aserto.com:8443"
  tenant_id: "<<<redacted>>>"
  directory_api_key: "<<<redacted>>>"

````

run

```
aserto-idp-sync run --config config.yaml
```

grpcurl

```
grpcurl -insecure -d '{"email_address": "gert.drapers@live.com"}' localhost:8282 idpsync.v1.IDPSync.SyncUser
```


curl

```
curl -k -d '{"email_address": "gert.drapers@live.com"}' -X POST https://localhost:8383/api/v1/sync/user
```

