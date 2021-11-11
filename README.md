# aserto-idp-sync

## Aserto IDP Sync 

`aserto-idp-sync` is a specialized service that enables the blocking synchronization of an IDP user, into the Aserto directory. This capability is especially useful when implementing sign-up flows. 

NOTE: Currently, this service only supports synchronizing Auth0 managed users.

The services can be run inside your environment and is explicitly configured for your tenant, using your aserto-tenant-id and an API key to communicate to the Aserto directory. 

The Auth0 configuration requires a domain, client-id, and client-secret; similar to how you configure the Auth0 IDP connection inside Aserto, the Auth0 connection used by the service only requires the `read:users` scope.


## Configuration

The service is configured using a yaml based configuration file, passed as a command line arrugment to the service, for example: `-c config.yaml` or `--config config.yaml`

### config.yaml

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

```

### Running the service

To run the service using a local binary:

```
aserto-idp-sync run --config config.yaml
```

To run the service using a Docker container image:


```
docker run -ti \
--platform=linux/amd64 \
--name aserto-idp-sync \
--rm \
-p 8282:8282 \
-p 8383:8383 \
-p 8484:8484 \
-v $PWD:/cfg \
ghcr.io/aserto-dev/aserto-idp-sync:latest run --config=/cfg/config-dev.yaml
```

### Integration

You can interact with the service via REST or gRPC API calls. 


#### gRPC example using grpcurl

```
grpcurl -insecure -d '{"email_address": "gert.drapers@live.com"}' localhost:8282 idpsync.v1.IDPSync.SyncUser
```


#### http REST example using curl 

```
curl -k -d '{"email_address": "gert.drapers@live.com"}' -X POST https://localhost:8383/api/v1/sync/user
```

