yes, i use master branch not main :v (just because that tech drama went viral)

# Testing GitOps guide

## Repo layout (what lives where)

- `deploy/infra/base` + overlays: cloudflared tunnel, kgateway, metallb, rook,
  cnpg, harbor (still flaky), reflector (we ended up shipping pull secrets
  per-namespace instead).
- `deploy/apps/base`: namespaces, Services, Deployments, HTTPRoutes for `go-app`
  and `nodejs-app`. Toggle them via `deploy/apps/base/kustomization.yaml` when
  you want them live.
- `apps/go-app`, `apps/nodejs-app`: source and Containerfiles, both serve
  on 3000.
- Secrets: SOPS+age (`testing.agekey`, `.sops.yaml`). GHCR pull secrets are
  committed encrypted and applied by Flux.
- Cluster: k3s single node (control plane + worker).

## CI/CD (what happens on push/tag)

- `build.yaml` (push to master): matrix build for go-app/nodejs-app with Buildx;
  tags `test-dryrun` and `${SHORT_SHA}-dryrun`; no deploy.
- `release.yaml` (tags `go-app-*` / `nodejs-app-*`): build/push to GHCR
  (`ghcr.io/idiot29/<app>:{version,short_sha,latest}`), check out `master`, bump
  the deployment image, commit, and push to `master`.
- Deploy via Flux: uncomment app entries in
  `deploy/apps/base/kustomization.yaml` and uncomment the `image:` lines in each
  deployment once you’re ready.

## How to access

- go app: `https://go.oryzasa.site` → `/`, `/health`, `/metrics`.
- nodejs app: `https://node.oryzasa.site` (same paths). Traffic flows through
  kgateway + cloudflared.
- Harbor: `https://harbor.oryzasa.site` (still temperamental behind Cloudflare).
- Rook dashboard: via HTTPRoute, TLS disabled per current cluster setup.
- Monitoring/Grafana: `https://grafana.oryzasa.site` (currently stuck in login
  redirect loops behind Cloudflare; cookies don’t stick).

## Run locally with podman

- From repo root: `cd apps && podman compose -f compose.yaml up -d` (shared otel
  collector).
- Per app:
  - `cd apps/go-app && podman compose -f compose.yaml up -d --build`
  - `cd apps/nodejs-app && podman compose -f compose.yaml up -d --build`
- Local ports:
  - go-app: http://localhost:3002 (`/`, `/health`, `/metrics`)
  - nodejs-app: http://localhost:3001 (`/`, `/health`, `/metrics`)

## Known issues / gotchas

- Reflector: unreliable for now; pull secrets are provisioned per-namespace
  (`deploy/apps/base/namespaces/ghcr-pullsecrets.yaml`).
- Harbor: CF proxy/auth is shaky; pushing to Harbor may fail—use GHCR for now.
- Tracing: set `OTEL_EXPORTER_OTLP_ENDPOINT` to a clean URL (e.g.,
  `http://otel-collector:4318`) to stop noisy errors.
- Grafana: Cloudflare/KGateway causes login redirects (cookies not sticky).

## GitOps loop

1. Edit manifests (keep secrets encrypted with SOPS).
2. Commit to `master`.
3. Flux watches `deploy/clusters/testing-cluster`; force apply with
   `flux reconcile kustomization flux-system -n flux-system` if needed.

## TESTING GAN!
