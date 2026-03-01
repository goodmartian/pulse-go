#!/usr/bin/env bash
# Creates an isolated demo environment for pulse VHS recording.
# Usage: source demo/setup.sh

set -euo pipefail

DEMO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEMO_PROJECTS="$DEMO_DIR/projects"
DEMO_CONFIG="$DEMO_DIR/config"

# Clean previous run
rm -rf "$DEMO_PROJECTS" "$DEMO_CONFIG"
mkdir -p "$DEMO_PROJECTS/.pulse/projects" "$DEMO_CONFIG/pulse"

# --- Helper: create a fake git repo with N commits ---
make_repo() {
  local name="$1" count="$2"
  shift 2
  local dir="$DEMO_PROJECTS/$name"
  mkdir -p "$dir"
  git -C "$dir" init -q
  git -C "$dir" config user.email "demo@pulse.dev"
  git -C "$dir" config user.name "Demo User"

  for i in $(seq 1 "$count"); do
    local msg="${1:-update $name}"
    shift 2>/dev/null || true
    echo "// commit $i" >> "$dir/main.go"
    git -C "$dir" add -A
    git -C "$dir" commit -q -m "$msg" --date="$(date -d "-$((count - i)) hours" --iso-8601=seconds)"
  done
}

# --- Projects ---

# Active / focused project — today's commits
make_repo "my-awesome-api" 3 \
  "feat: add user authentication endpoint" \
  "fix: handle expired JWT tokens gracefully" \
  "test: add integration tests for auth flow"

# Hot projects
make_repo "dashboard-ui" 2 \
  "feat: add dark mode toggle" \
  "refactor: extract chart component"

make_repo "cli-toolkit" 1 \
  "feat: initial project structure"

# Warm project — commits a few days ago
dir="$DEMO_PROJECTS/mobile-app"
mkdir -p "$dir"
git -C "$dir" init -q
git -C "$dir" config user.email "demo@pulse.dev"
git -C "$dir" config user.name "Demo User"
echo "// v1" > "$dir/app.dart"
git -C "$dir" add -A
WARM_DATE="$(date -d '-5 days' --iso-8601=seconds)"
GIT_COMMITTER_DATE="$WARM_DATE" git -C "$dir" commit -q -m "feat: onboarding screen" --date="$WARM_DATE"

# Cold project — old commit
dir="$DEMO_PROJECTS/old-experiment"
mkdir -p "$dir"
git -C "$dir" init -q
git -C "$dir" config user.email "demo@pulse.dev"
git -C "$dir" config user.name "Demo User"
echo "# experiment" > "$dir/README.md"
git -C "$dir" add -A
COLD_DATE="$(date -d '-60 days' --iso-8601=seconds)"
GIT_COMMITTER_DATE="$COLD_DATE" git -C "$dir" commit -q -m "Initial commit" --date="$COLD_DATE"

# --- Project metadata ---
cat > "$DEMO_PROJECTS/.pulse/projects/my-awesome-api.json" << 'EOF'
{
  "name": "my-awesome-api",
  "tagline": "REST API for the main product",
  "description": "Backend service: auth, billing, notifications",
  "done_when": "All endpoints covered with tests, deployed to staging",
  "stack": "Go, PostgreSQL, Redis",
  "notes": ""
}
EOF

cat > "$DEMO_PROJECTS/.pulse/projects/dashboard-ui.json" << 'EOF'
{
  "name": "dashboard-ui",
  "tagline": "Admin dashboard",
  "description": "React dashboard for managing users and analytics",
  "done_when": "Dark mode + charts + responsive layout",
  "stack": "React, TypeScript, TailwindCSS",
  "notes": ""
}
EOF

cat > "$DEMO_PROJECTS/.pulse/projects/cli-toolkit.json" << 'EOF'
{
  "name": "cli-toolkit",
  "tagline": "Internal CLI tools",
  "description": "Developer tooling for code generation and deploys",
  "done_when": "Core generators working",
  "stack": "Go, Cobra",
  "notes": ""
}
EOF

# --- State: focus + statuses ---
cat > "$DEMO_CONFIG/pulse/state.json" << 'EOF'
{
  "focus": "my-awesome-api",
  "focus_since": "2026-02-27",
  "statuses": {
    "my-awesome-api": "active",
    "dashboard-ui": "active",
    "cli-toolkit": "active",
    "mobile-app": "active",
    "old-experiment": "paused"
  },
  "switches": [
    {"from": "cli-toolkit", "to": "mobile-app", "reason": "user testing next week", "date": "2026-02-20"},
    {"from": "mobile-app", "to": "dashboard-ui", "reason": "client demo on Friday", "date": "2026-02-24"},
    {"from": "dashboard-ui", "to": "my-awesome-api", "reason": "deadline approaching", "date": "2026-02-27"}
  ]
}
EOF

# --- Ideas ---
cat > "$DEMO_CONFIG/pulse/ideas.json" << 'EOF'
[
  {"text": "Add WebSocket support for real-time updates", "date": "2026-02-28"},
  {"text": "Try Bun instead of Node for dashboard", "date": "2026-02-27"},
  {"text": "Write a blog post about Go concurrency patterns", "date": "2026-02-26"}
]
EOF

# --- Config pointing to demo projects ---
cat > "$DEMO_CONFIG/pulse/config.json" << EOF
{
  "projects_dir": "$DEMO_PROJECTS",
  "lang": "en"
}
EOF

# --- Export env for pulse ---
export XDG_CONFIG_HOME="$DEMO_CONFIG"
export PULSE_DIR="$DEMO_PROJECTS"

echo "✓ Demo environment ready"
echo "  Projects: $DEMO_PROJECTS"
echo "  Config:   $DEMO_CONFIG/pulse"
echo "  Run: ./pulse"
