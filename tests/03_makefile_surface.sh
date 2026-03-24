#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
services=(wifi-probe dns-probe jitter-probe gateway-monitor alert-receiver)

required_make_vars=(
  "IMAGE_TAG"
  "REGISTRY"
  "K3D_CLUSTER"
  "K3S_REGISTRY"
  "KUBE_CONTEXT"
)

required_make_targets=(
  "build-image"
  "push-k3d"
  "push"
  "deploy"
  "deploy-k3s"
  "rollout"
  "require-kube-context"
)

for svc in "${services[@]}"; do
  makefile="$ROOT_DIR/$svc/Makefile"

  for var in "${required_make_vars[@]}"; do
    grep -q "$var" "$makefile" || {
      printf "%s Makefile missing variable hint: %s\n" "$svc" "$var" >&2
      exit 1
    }
  done

  for target in "${required_make_targets[@]}"; do
    grep -Eq "^\\.PHONY: ${target}$|^\\.PHONY:.*\\b${target}\\b|^${target}:" "$makefile" || {
      printf "%s Makefile missing target: %s\n" "$svc" "$target" >&2
      exit 1
    }
  done
done

printf "Makefile surface checks passed.\n"
