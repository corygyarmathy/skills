#!/usr/bin/env bash
set -euo pipefail

# Links every skill in this repo into ~/.claude/skills, so an edit here is live
# in the next Claude Code session with no copy step. OpenCode reads the same
# directory. Re-run it after adding, renaming or deleting a skill: links whose
# skill is gone are removed, so a rename doesn't leave the old name behind.
#
# This is for iterating on the skills. A consumer that must work without this
# checkout (afk-agent) vendors a pinned copy instead; see the README.

REPO="$(cd "$(dirname "$0")/.." && pwd)"
DEST="${CLAUDE_SKILLS_DIR:-$HOME/.claude/skills}"

# A $DEST that resolves into this repo would put the links inside skills/.
case "$(readlink -f "$DEST" 2>/dev/null || true)" in
"$REPO" | "$REPO"/*)
    echo "error: $DEST resolves into this repo; remove it and re-run." >&2
    exit 1
    ;;
esac

mkdir -p "$DEST"

# Drop links into this repo whose skill no longer exists.
for link in "$DEST"/*; do
    [ -L "$link" ] || continue
    target="$(readlink "$link")"
    case "$target" in
    "$REPO"/skills/*)
        if [ ! -f "$target/SKILL.md" ]; then
            rm "$link"
            echo "removed $(basename "$link") (no longer in this repo)"
        fi
        ;;
    esac
done

for skill_md in "$REPO"/skills/*/SKILL.md; do
    src="$(dirname "$skill_md")"
    name="$(basename "$src")"
    target="$DEST/$name"

    # Never delete a real directory: it may be a skill installed some other way.
    if [ -e "$target" ] && [ ! -L "$target" ]; then
        echo "skipped $name: $target exists and is not a symlink" >&2
        continue
    fi

    ln -sfn "$src" "$target"
    echo "linked $name"
done
