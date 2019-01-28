#!/usr/bin/env bash

set -euo pipefail

cd "$( dirname "$0" )"

mkdir -p mocks

mockgen -destination=mocks/mock_dbi.go -package=mocks github.com/merlincox/cardapi/db Dbi

