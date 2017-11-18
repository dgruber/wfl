#!/bin/sh

# create a random string of length specified in arg[0] and save it in arg[1]
env LC_CTYPE=C tr -dc 'a-z' < /dev/urandom | head -c$1 > $2 
