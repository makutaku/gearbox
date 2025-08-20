#!/bin/bash
echo "MOCK_SCRIPT_ARGS_COUNT: $#" >&2
echo "MOCK_SCRIPT_ARGS: $*" >&2
for i in "$@"; do
    echo "MOCK_SCRIPT_ARG: $i" >&2
done
exit 0
