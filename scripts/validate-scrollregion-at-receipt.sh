#!/usr/bin/env bash
# Validate a manually captured final AT receipt. This script never captures or
# fabricates AT evidence: it only invokes the fail-closed Go verifier below.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
receipt_path="${GOSHTOSO_SCROLLREGION_AT_RECEIPT:-}"
identity_path="${GOSHTOSO_SCROLLREGION_AT_IDENTITY:-}"
challenge_path="${GOSHTOSO_SCROLLREGION_AT_CHALLENGE:-}"
registry_path="${GOSHTOSO_SCROLLREGION_AT_REPLAY_REGISTRY:-}"

if [[ -z "$receipt_path" || -z "$identity_path" || -z "$challenge_path" || -z "$registry_path" ]]; then
  cat >&2 <<'EOF'
Final T-GS-011 AT capture is external/UI-bound. Supply a real capture and the
frozen candidate identity file; the checked-in template is intentionally not
evidence:

  GOSHTOSO_SCROLLREGION_AT_RECEIPT=/absolute/final-at-receipt.json \
  GOSHTOSO_SCROLLREGION_AT_IDENTITY=/absolute/frozen-candidate-identity.json \
  GOSHTOSO_SCROLLREGION_AT_CHALLENGE=/absolute/independent-at-challenge.json \
  GOSHTOSO_SCROLLREGION_AT_REPLAY_REGISTRY=/absolute/owner-only-challenge-registry \
  scripts/validate-scrollregion-at-receipt.sh
EOF
  exit 2
fi
[[ -r "$receipt_path" ]] || { echo "AT receipt is not readable: $receipt_path" >&2; exit 1; }
[[ -r "$identity_path" ]] || { echo "AT identity is not readable: $identity_path" >&2; exit 1; }
[[ -r "$challenge_path" ]] || { echo "AT challenge is not readable: $challenge_path" >&2; exit 1; }
[[ -d "$registry_path" ]] || { echo "AT replay registry is not a directory: $registry_path" >&2; exit 1; }

cd "$repo_root"
GOSHTOSO_SCROLLREGION_AT_RECEIPT="$receipt_path" \
GOSHTOSO_SCROLLREGION_AT_IDENTITY="$identity_path" \
GOSHTOSO_SCROLLREGION_AT_CHALLENGE="$challenge_path" \
GOSHTOSO_SCROLLREGION_AT_REPLAY_REGISTRY="$registry_path" \
go test -tags='e2e,scrollregion,bfull,axe' ./site/tests/e2e \
  -count=1 -timeout=5m -run '^TestScrollRegionBFullATReceiptHarness$'
