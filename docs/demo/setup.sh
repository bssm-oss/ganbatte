#!/usr/bin/env bash
# Pre-create demo fixture environments for vhs tapes.
set -e

# --- migrate demo ---
MHOME=/tmp/gnbd-migrate
rm -rf "$MHOME"
mkdir -p "$MHOME/.config/ganbatte"
printf 'version = "1.0.0"\n' > "$MHOME/.config/ganbatte/config.toml"
cat > "$MHOME/.zshrc" << 'EOF'
alias gs='git status -sb'
alias gl='git log --oneline -10'
alias ll='eza -alF --git'
alias dc='docker compose'
alias pdev='pnpm run dev'
alias pbuild='pnpm run build'
alias serve='python3 -m http.server'
alias cdp='cd ~/Projects'
EOF
echo "migrate env ready: $MHOME"

# --- suggest demo ---
SHOME=/tmp/gnbd-suggest
rm -rf "$SHOME"
mkdir -p "$SHOME/.config/ganbatte"
printf 'version = "1.0.0"\n' > "$SHOME/.config/ganbatte/config.toml"
python3 - << 'PYEOF'
import time, os
base = int(time.time()) - 7 * 86400
lines = []
for i in range(6):
    lines.append(f': {base + i * 3600}:0;claude')
repos = ['github.com/user/repo-a', 'github.com/org/project', 'github.com/foo/bar',
         'github.com/bssm/tool', 'github.com/user/another']
for j in range(3):
    for i, r in enumerate(repos if j < 2 else repos[:3]):
        t = base + 86400*(j+1) + i * 7200
        lines.append(f': {t}:0;git clone https://{r}.git')
        lines.append(f': {t+60}:0;ls')
pkgs = ['typescript@5', 'prettier@3', '@anthropic-ai/claude-code', 'pnpm@9', 'bun@1.1']
for i, p in enumerate(pkgs * 2):
    lines.append(f': {base + 345600 + i * 3600}:0;npm i -g {p}')
for i in range(7):
    t = base + 432000 + i * 90000
    lines.append(f': {t}:0;git add .')
    lines.append(f': {t+30}:0;git commit -m "fix: update"')
    lines.append(f': {t+60}:0;git push')
with open('/tmp/gnbd-suggest/.zsh_history', 'w') as f:
    f.write('\n'.join(lines) + '\n')
PYEOF
echo "suggest env ready: $SHOME"

# --- tui demo ---
THOME=/tmp/gnbd-tui
rm -rf "$THOME"
mkdir -p "$THOME/.config/ganbatte"
cat > "$THOME/.config/ganbatte/config.toml" << 'EOF'
version = "1.0.0"

[alias.gs]
cmd = "git status -sb"

[alias.gl]
cmd = "git log --oneline -10"

[alias.gco]
cmd = "git checkout {branch}"
params = ["branch"]

[alias.dc]
cmd = "docker compose"

[alias.serve]
cmd = "python3 -m http.server"

[alias.nuke]
cmd = "git reset --hard HEAD"
confirm = true
tags = ["git", "danger"]

[workflow.deploy]
description = "Lint, test, build, push"
params = ["branch"]
steps = [
  { run = "pnpm lint" },
  { run = "pnpm test", on_fail = "stop" },
  { run = "pnpm build" },
  { run = "git push origin {branch}", confirm = true },
]
tags = ["deploy", "ci"]

[workflow.setup]
description = "Bootstrap project"
steps = [
  { run = "npm install" },
  { run = "cp .env.example .env" },
]
tags = ["onboarding"]
EOF
echo "tui env ready: $THOME"
