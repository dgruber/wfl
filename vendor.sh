#!/bin/sh

rm -rf vendor/ && govendor init && govendor add +e +x github.com/dgruber/drmaa2interface
rm -rf vendor/github.com/dgruber/drmaa2interface

