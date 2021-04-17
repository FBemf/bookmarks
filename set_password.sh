#!/usr/bin/env bash

set -euo pipefail

USERNAME="$1"
printf 'Setting password for user %s\n' "$USERNAME"
IFS=
read -serp "Password: " PASSWORD
printf '\n'
./bookmarks user -username "$USERNAME" -password "$PASSWORD"