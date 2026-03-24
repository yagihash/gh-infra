#!/usr/bin/env bash
# Setup script for VHS demo recording.
# Runs inside the VHS Docker container during the Hide phase.
#
# Expects:
#   /data/.gh-infra  — pre-built Linux binary (built by 'make demos')
#   /data/mock-gh    — mock gh CLI with canned API responses
set -euo pipefail

# Install pre-built gh-infra
cp /data/.gh-infra /usr/local/bin/gh-infra
chmod +x /usr/local/bin/gh-infra

# Install gh wrapper:
#   'gh infra ...' → real gh-infra binary
#   everything else → mock-gh (canned API responses with delay)
cat > /usr/local/bin/gh << 'WRAPPER'
#!/usr/bin/env bash
if [[ "$1" == "infra" ]]; then
  shift
  exec /usr/local/bin/gh-infra "$@"
fi
exec /data/mock-gh "$@"
WRAPPER
chmod +x /usr/local/bin/gh

# Create demo YAML
mkdir -p /tmp/demo
cat > /tmp/demo/repos.yaml << 'EOF'
apiVersion: gh-infra/v1
kind: Repository
metadata:
  name: my-project
  owner: babarot

spec:
  description: "My awesome project"
  visibility: public
  topics:
    - go
    - cli
  features:
    issues: true
    wiki: false
    discussions: true
  merge_strategy:
    allow_squash_merge: true
    auto_delete_head_branches: true
EOF

# Pass through env vars from host (e.g. GH_INFRA_OUTPUT=stream)
export GH_INFRA_OUTPUT="${GH_INFRA_OUTPUT:-}"

export PS1='$ '
