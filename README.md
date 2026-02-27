# cert-manager-webhook-inwx

A [cert-manager](https://cert-manager.io/) DNS01 webhook solver for [INWX](https://www.inwx.de/). This webhook allows cert-manager to automatically create and clean up DNS TXT records via the INWX API to solve ACME DNS-01 challenges, enabling automatic TLS certificate issuance for domains managed by INWX.

## How It Works

When cert-manager needs to verify domain ownership for a TLS certificate, it uses the ACME DNS-01 challenge type. This webhook handles two operations:

1. **Present** — Creates a `_acme-challenge.{domain}` TXT record in INWX with the challenge token
2. **CleanUp** — Deletes the specific TXT record after validation completes

The webhook runs as a Kubernetes API extension server. cert-manager sends `ChallengeRequest` resources to it, and the webhook translates those into INWX API calls. Multiple concurrent challenges for the same domain are supported — cleanup only removes the record matching the specific challenge key.

## Architecture

```
cert-manager ──► APIService ──► webhook pod ──► INWX API
                 (v1alpha1)     (this project)   (DNS records)
```

The webhook registers itself as a Kubernetes API extension under the group `inwx.webhook.cert-manager.io`. cert-manager discovers it via the `APIService` resource and sends challenge requests over HTTPS. The webhook authenticates with the INWX XML-RPC API using username/password credentials.

## Libraries

- [nrdcg/goinwx](https://github.com/nrdcg/goinwx) — Go client for the INWX domain API
- [cert-manager/cert-manager](https://github.com/cert-manager/cert-manager) — ACME webhook SDK (`pkg/acme/webhook`)
- [k8s.io/apiextensions-apiserver](https://github.com/kubernetes/apiextensions-apiserver) — Kubernetes API extension framework

## Prerequisites

- A Kubernetes cluster with [cert-manager](https://cert-manager.io/docs/installation/) installed
- An [INWX](https://www.inwx.de/) account with API access
- Helm 3 (for installation)

## Installation

### Using Helm

```bash
helm install cert-manager-webhook-inwx \
  ./deploy/cert-manager-webhook-inwx \
  --namespace cert-manager \
  --set inwx.username=YOUR_INWX_USERNAME \
  --set inwx.password=YOUR_INWX_PASSWORD
```

Or with an existing Kubernetes secret containing your credentials:

```bash
# Create the secret first
kubectl create secret generic inwx-credentials \
  --namespace cert-manager \
  --from-literal=username=YOUR_INWX_USERNAME \
  --from-literal=password=YOUR_INWX_PASSWORD

# Install with existing secret reference
helm install cert-manager-webhook-inwx \
  ./deploy/cert-manager-webhook-inwx \
  --namespace cert-manager \
  --set inwx.existingSecret=inwx-credentials
```

### Helm Values

| Parameter | Description | Default |
|---|---|---|
| `groupName` | API group name for the webhook | `inwx.webhook.cert-manager.io` |
| `image.repository` | Container image repository | `ghcr.io/antimatter-studios/cert-manager-webhook-inwx` |
| `image.tag` | Container image tag | `latest` |
| `inwx.username` | INWX API username | `""` |
| `inwx.password` | INWX API password | `""` |
| `inwx.sandbox` | Use INWX sandbox environment | `false` |
| `inwx.existingSecret` | Name of existing secret with `username` and `password` keys | `""` |
| `certManager.namespace` | Namespace where cert-manager is installed | `cert-manager` |
| `certManager.serviceAccountName` | cert-manager service account name | `cert-manager` |

## Usage

### 1. Create a ClusterIssuer (or Issuer)

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod-account-key
    solvers:
    - dns01:
        webhook:
          groupName: inwx.webhook.cert-manager.io
          solverName: inwx
```

### 2. Request a Certificate

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example-com-tls
  namespace: default
spec:
  secretName: example-com-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
  - example.com
  - "*.example.com"
```

cert-manager will automatically use the INWX webhook to create the `_acme-challenge` TXT records needed for DNS-01 validation, then clean them up after the certificate is issued.

### Wildcard Certificates

DNS-01 is the only ACME challenge type that supports wildcard certificates. This webhook enables that — just include `*.yourdomain.com` in the `dnsNames` list.

## Development

### Building

```bash
# Build locally
make build

# Run tests
make test

# Build Docker image
make docker-build
```

### Testing

The project includes unit tests covering the core solver logic:

```bash
go test -v ./inwx
```

Tests use a mock INWX client to verify:
- TXT record creation (Present)
- TXT record deletion (CleanUp)
- Concurrent challenge support (only matching records are deleted)
- Idempotent Present calls
- Subdomain handling

### Environment Variables

| Variable | Description | Required |
|---|---|---|
| `GROUP_NAME` | Kubernetes API group name | Yes |
| `INWX_USERNAME` | INWX account username | Yes |
| `INWX_PASSWORD` | INWX account password | Yes |
| `INWX_SANDBOX` | Use INWX sandbox (`true`/`false`) | No (default: `false`) |

## Container Image

Multi-architecture images (linux/amd64, linux/arm64) are published to GitHub Container Registry:

```
ghcr.io/antimatter-studios/cert-manager-webhook-inwx
```

Images are built automatically on pushes to `main` and on version tags (`v*`).

## Related Projects

- [external-dns-webhook-inwx](https://github.com/antimatter-studios/external-dns-webhook-inwx) — ExternalDNS webhook provider for INWX (manages A, CNAME, and other DNS records)

## License

See [LICENSE](LICENSE) for details.
