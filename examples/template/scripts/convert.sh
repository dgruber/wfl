#!/bin/sh

# convert n't char of input to uppercase
input=`head -c $TASK_ID $1 | tail -c 1 | awk '{ print toupper($0) }'`
echo ${input} > ./output/${TASK_ID}.txt
