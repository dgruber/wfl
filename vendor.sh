#!/bin/sh

rm -rf vendor/ && govendor init && govendor add +e
rm -rf vendor/github.com/dgruber/drmaa2interface

