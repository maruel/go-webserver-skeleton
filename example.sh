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
req -d '{"name":"stdout"}' http://localhost:8081/api/log/manual
req -d '{"name":"stderr"}' http://localhost:8081/api/log/manual
# Handler refuses this request:
req -d '{}' http://localhost:8081/api/log/manual
# GET is disallowed:
req http://localhost:8081/api/log/manual

echo "- With reflection based API:"
req -d '{"name":"stdout"}' http://localhost:8081/api/log/auto
req -d '{"name":"stderr"}' http://localhost:8081/api/log/auto
# Handler refuses this request:
req -d '{}' http://localhost:8081/api/log/auto
# GET is disallowed:
req http://localhost:8081/api/log/auto

echo "- Quitting"
curl http://localhost:8081/quitquitquit
