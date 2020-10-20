#!/bin/bash
# Copyright 2020 Marc-Antoine Ruel. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

set -eu

exec 3>&1

function req {
  echo "  $@"
  curl -s -w "  HTTP %{http_code}\n" "$@" -o >(cat >&3)
  echo ""
  echo ""
}

echo "- Manual marshaling:"
URL=http://localhost:8081/api/log/manual
req -H "Content-Type: application/json" -d '{"name":"stdout"}' $URL
req -H "Content-Type: application/json" -d '{"name":"stderr"}' $URL
echo "  Handler refuses this request:"
req -H "Content-Type: application/json" -d '{"foo":"bar"}' $URL
echo "  GET is disallowed:"
req -H "Content-Type: application/json" $URL
echo "  Incorrect Content-Type:"
req -H "Content-Type: application/octet-steam" -d '{"name":"stderr"}' $URL

echo "- With reflection based API:"
URL=http://localhost:8081/api/log/auto
req -H "Content-Type: application/json" -d '{"name":"stdout"}' $URL
req -H "Content-Type: application/json" -d '{"name":"stderr"}' $URL
echo "  Handler refuses this request:"
req -H "Content-Type: application/json" -d '{"foo":"bar"}' $URL
echo "  GET is disallowed:"
req -H "Content-Type: application/json" $URL
echo "  Incorrect Content-Type:"
req -H "Content-Type: application/octet-steam" -d '{"name":"stderr"}' $URL

echo "- Quitting"
curl http://localhost:8081/quitquitquit
