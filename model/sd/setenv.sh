#!/bin/bash
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi
echo "done"