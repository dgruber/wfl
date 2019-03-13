#!/bin/sh

rm -rf vendor/ && govendor init && govendor add +e
# don't vendor commonly used interfaces
rm -rf vendor/dgruber/drmaa2interface

