#!/bin/sh
fly -t ci set-pipeline -p wfl -c ./pipeline.yml

