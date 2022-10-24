#!/bin/env bash

# Run CI tests
CI_CMD="make ci"
eval $CI_CMD
RESULT=$?
if [ $RESULT -ne 0 ]; then
    echo -e "CI tests Failed!\n CMD: $CI_CMD"
    exit 1
fi