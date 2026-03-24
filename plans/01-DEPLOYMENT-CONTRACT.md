# Deterministic Deployment Contract

## Objective

Define a repeatable release contract so local k3d deployments and Raspberry Pi k3s deployments use the same release identity and verification criteria.

## Required Outcome

- one immutable `RELEASE_ID` used for every service in a deploy run
- fixed service deployment order
- explicit target selection: `k3d` or `k3s`
- explicit Kubernetes context for every `kubectl` and `helm` command
- explicit Helm values profile per target
- deterministic verification commands with pass/fail output

## Constraints

- do not deploy mutable tags (`latest`) for shared environments
- do not use per-service ad-hoc tags in the same release
- do not edit infrastructure definitions in this repository
- deployment steps must work by rerunning the same commands with the same inputs
- all workflow changes must keep `./tests/run-all.sh` passing

## Canonical Services

Deploy in this order for every release:

1. `wifi-probe`
2. `dns-probe`
3. `jitter-probe`
4. `gateway-monitor`
5. `alert-receiver`

`hello-world` is intentionally excluded from the production deployment set.

## Release Identity

Use one release ID per run:

```bash
export RELEASE_ID="$(date +%Y%m%d-%H%M)-$(git rev-parse --short HEAD)"
```

Rules:

- a release run reuses exactly one `RELEASE_ID`
- all charts are deployed with that same tag
- re-run with same `RELEASE_ID` means same intended state

## Required Local Tools

```bash
go version
docker version
helm version
kubectl version --client
k3d version   # required for local target only
```

## Baseline Validation

```bash
./tests/run-all.sh
```

Proceed to target-specific plans only after baseline checks pass.

## Helm Target Profiles

- k3d profile: `<service>/charts/<service>/values.yaml`
- k3s profile: `<service>/charts/<service>/values-k3s.yaml`

Both profiles are required for every canonical service.
