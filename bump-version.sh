#!/bin/bash
set -e

### Determine new version

# Fetch all remote branches to ensure we have access to origin/main
git fetch --all --prune

# Find the common ancestor (merge-base) of origin/main and HEAD
BASE_MAIN=$(git merge-base origin/main HEAD)

# Look for the latest commit that touched the VERSION file after BASE_MAIN
LAST_VERSION_COMMIT=$(git log --pretty=format:%H --since="$(git show -s --format=%ci $BASE_MAIN)" -- VERSION | head -n 1)

# Decide base commit
if [ -n "$LAST_VERSION_COMMIT" ]; then
  BASE="$LAST_VERSION_COMMIT"
else
  BASE="$BASE_MAIN"
fi

# Get all commit messages since that base
COMMITS=$(git log $BASE..HEAD --pretty=format:"%s%n%b")


# Read current version or default
if [ -f VERSION ]; then
  CURRENT_VERSION=$(cat VERSION)
else
  CURRENT_VERSION="0.0.0"
fi

IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"

# Determine version bump
if echo "$COMMITS" | grep -q "BREAKING CHANGE:"; then
  MAJOR=$((MAJOR + 1))
  MINOR=0
  PATCH=0
elif echo "$COMMITS" | grep -q "^feat:"; then
  MINOR=$((MINOR + 1))
  PATCH=0
elif echo "$COMMITS" | grep -q "^fix:"; then
  PATCH=$((PATCH + 1))
else
  # No relevant commit. Leave version unchanged
  echo "No versionable commits found"
  touch .bump-ignore
  exit 0
fi

NEW_VERSION="$MAJOR.$MINOR.$PATCH"

echo "$NEW_VERSION" > VERSION
echo "Bumped version to $NEW_VERSION"
[ -f .bump-ignore ] && rm .bump-ignore

# Generate CHANGELOG section
CHANGELOG_FILE="CHANGELOG.md"
CHANGELOG_SECTION="## [$NEW_VERSION] - $(date +%Y-%m-%d)\n"

ADDED=$(echo "$COMMITS" | grep "^feat:" | sed 's/^feat: */- /')
FIXED=$(echo "$COMMITS" | grep "^fix:" | sed 's/^fix: */- /')
CHANGED=$(echo "$COMMITS" | grep -E "^chore:|^refactor:" | sed 's/^\(chore\|refactor\): */- /')
REMOVED=$(echo "$COMMITS" | grep "^remove:" | sed 's/^remove: */- /')

if [ -n "$ADDED" ]; then
  CHANGELOG_SECTION+="\n### Added\n\n$ADDED\n"
fi

if [ -n "$FIXED" ]; then
  CHANGELOG_SECTION+="\n### Fixed\n\n$FIXED\n"
fi

if [ -n "$CHANGED" ]; then
  CHANGELOG_SECTION+="\n### Changed\n\n$CHANGED\n"
fi

if [ -n "$REMOVED" ]; then
  CHANGELOG_SECTION+="\n### Removed\n\n$REMOVED\n"
fi

# Prepend to CHANGELOG.md
CHANGELOG_HEADER="# Changelog"

if [ -f "$CHANGELOG_FILE" ]; then
  # Separate the header and the rest
  EXISTING_CONTENT=$(tail -n +2 "$CHANGELOG_FILE")
  echo -e "$CHANGELOG_HEADER\n\n$CHANGELOG_SECTION$EXISTING_CONTENT" > $CHANGELOG_FILE
else
  echo -e "$CHANGELOG_HEADER\n\n$CHANGELOG_SECTION" > $CHANGELOG_FILE
fi

echo "Changelog updated"