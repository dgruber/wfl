#!/bin/sh

output=""
for i in `seq 0 $1`; do
    file=`cat ./output/$i.txt`
    output=$output$file
done

echo ${output} > $2

