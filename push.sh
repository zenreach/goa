#!/bin/bash -ex

godep go test
godocdown . > ./README.md
branch_name="$(git symbolic-ref HEAD 2>/dev/null)" || branch_name="(unnamed branch)"
branch_name=${branch_name##refs/heads/}
git push origin $branch_name

