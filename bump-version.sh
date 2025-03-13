#!/bin/bash

set -e

VERSION_FILE="version/version.go"
README_FILE="README.md"
CURRENT_VERSION=$(grep 'const Version' $VERSION_FILE | awk -F '"' '{print $2}')

if [[ -z "$CURRENT_VERSION" ]]; then
    echo "Error: Could not find the current version in $VERSION_FILE"
    exit 1
fi

echo "Current version: $CURRENT_VERSION"


# Determine bump type
if [[ "$1" == "major" ]]; then
    NEW_VERSION=$(echo $CURRENT_VERSION | awk -F. '{print $1+1 ".0.0"}')
elif [[ "$1" == "minor" ]]; then
    NEW_VERSION=$(echo $CURRENT_VERSION | awk -F. '{print $1 "." $2+1 ".0"}')
elif [[ "$1" == "patch" ]]; then
    NEW_VERSION=$(echo $CURRENT_VERSION | awk -F. '{print $1 "." $2 "." $3+1}')
else
    echo "Error: Unknown bump type. Use 'major', 'minor', or 'patch'."
    exit 1
fi

echo "New version: $NEW_VERSION"

# Update version.go
sed -i "s/$CURRENT_VERSION/$NEW_VERSION/g" $VERSION_FILE

# Update README.md (format "Current Version: **vX.Y.Z**"))
sed -i "s/Current version: \*\*v$CURRENT_VERSION\*\*/Current version: **v$NEW_VERSION**/g" "$README_FILE"


echo "Version bumped to $NEW_VERSION."
