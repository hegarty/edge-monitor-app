#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

required_agent_refs=(
  "CLAUDE.md"
  "plans/00-START.md"
  "plans/02-K3D-DEPLOYMENT.md"
  "plans/03-K3S-DEPLOYMENT.md"
  "plans/04-K3S-REMOTE-WRITE.md"
  "tests/TESTS.md"
)

for ref in "${required_agent_refs[@]}"; do
  grep -qF "$ref" "$ROOT_DIR/AGENTS.md" || {
    printf "AGENTS.md is missing required reference: %s\n" "$ref" >&2
    exit 1
  }
done

required_plan_order=(
  "01-DEPLOYMENT-CONTRACT.md"
  "02-K3D-DEPLOYMENT.md"
  "03-K3S-DEPLOYMENT.md"
  "04-K3S-REMOTE-WRITE.md"
)

for plan in "${required_plan_order[@]}"; do
  grep -qF "$plan" "$ROOT_DIR/plans/00-START.md" || {
    printf "plans/00-START.md is missing plan reference: %s\n" "$plan" >&2
    exit 1
  }
done

required_k3s_profile_refs=(
  "values-k3s.yaml"
  "deploy-k3s"
)

for ref in "${required_k3s_profile_refs[@]}"; do
  grep -qF "$ref" "$ROOT_DIR/plans/03-K3S-DEPLOYMENT.md" || {
    printf "plans/03-K3S-DEPLOYMENT.md is missing required reference: %s\n" "$ref" >&2
    exit 1
  }
done

required_claude_sections=(
  "# Deterministic Deployment Contract"
  "# Release Identity Policy"
  "# Target Profiles"
  "# How To Operate This Repository"
  "# Verification Workflow"
)

for section in "${required_claude_sections[@]}"; do
  grep -qF "$section" "$ROOT_DIR/CLAUDE.md" || {
    printf "CLAUDE.md is missing required section: %s\n" "$section" >&2
    exit 1
  }
done

printf "Documentation consistency checks passed.\n"
