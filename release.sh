#!/usr/bin/env bash

set -o xtrace
set -o errexit
set -o nounset
set -o pipefail

if [ ! -d ./.git ]
then
  echo "not in root directory"
  exit 1
fi

if [ ! -z "$(git status --porcelain)" ]
then
  echo "working tree is not clean"
  exit 1
fi

aws sts get-caller-identity

git checkout master
git pull

VERSION="$(cat ./common.mk | egrep '^VERSION=')"
# check that the version is formatted as we expect
echo "$VERSION" | egrep '^VERSION=[0-9]+\.[0-9]+\.[0-9]+-SNAPSHOT$'
VERSION="$(echo "$VERSION" | gcut -d "=" -f 2)"
VERSION="$(echo "$VERSION" | gcut -d "-" -f 1)"
VERSION_MAJOR="$(echo "$VERSION" | gcut -d "." -f 1)"
VERSION_MINOR="$(echo "$VERSION" | gcut -d "." -f 2)"
VERSION_PATCH="$(echo "$VERSION" | gcut -d "." -f 3)"
function set_version()
{
  gsed -i -e 's/^VERSION=.*$/VERSION='"$1"'/' ./common.mk
}

if [ "$VERSION_PATCH" != "0" ]
then
  echo "patchlevel is not zero"
  exit 1
fi

VERSION_THIS="$VERSION_MAJOR"."$VERSION_MINOR"."$VERSION_PATCH"
VERSION_NEXT="$VERSION_MAJOR"."$(( (VERSION_MINOR + 1) ))".0-SNAPSHOT

git checkout -b                releases/substrateplugin/"$VERSION_THIS"
git push --set-upstream origin releases/substrateplugin/"$VERSION_THIS"

set_version "$VERSION_THIS"

make clean
make
make citest

git commit -a -m 'Create release version '"$VERSION_THIS"
git tag -a -f -m 'Release '"$VERSION_THIS" v"$VERSION_THIS"

make publish

set_version "$VERSION_NEXT"

make

git commit -a -m 'Set version to '"$VERSION_NEXT"

set +o xtrace
echo "Remember, you must still push tags, push branch, create pull request, and change branches ..."
echo "+OK (release.sh)"
