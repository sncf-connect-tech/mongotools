#!/bin/bash

NEW_VERSION=$1
CURRENT_VERSION=`cat VERSION`

CURRENT_TAG="v${CURRENT_VERSION}"

echo "release tag $CURRENT_TAG"

git tag -a $CURRENT_TAG -m "release $CURRENT_TAG"
git push --tags

START_HASH=`git rev-list --tags --skip=1 --max-count=1`
START_TAG=`git describe --abbrev=0 --tags $START_HASH`

echo "changelog since tag $START_TAG"

git log --oneline --decorate $START_TAG..$CURRENT_TAG 
git log --oneline --decorate $START_TAG..$CURRENT_TAG > /tmp/mongoanonymize_gitlog

echo "# release [$CURRENT_TAG](http://nexus/content/repositories/dt-releases/com/vsct/dt/mongoanonymize/$VERSION/)" >> /tmp/mongoanonymize_changelog.md
echo "" >> /tmp/mongoanonymize_changelog.md

# update changelog
while read -r line
do
  r=`echo $line | sed 's/^\([a-z0-9]*\)/[\1](http:\/\/gitlab.socrate.vsct.fr\/dt\/mongoanonymize\/commit\/\1)/' `
  echo "* $r" >> /tmp/mongoanonymize_changelog.md
done < "/tmp/mongoanonymize_gitlog"

echo "" >> /tmp/mongoanonymize_changelog.md

cat CHANGELOG.md >> /tmp/mongoanonymize_changelog.md
mv /tmp/mongoanonymize_changelog.md CHANGELOG.md

# update VERSION file
echo $NEW_VERSION > VERSION

# commit&push changed files
git add VERSION
git add CHANGELOG.md
git commit -m "CHANGELOG $CURRENT_TAG"
git push origin master
