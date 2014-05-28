#!/bin/bash

set -e

dir="instances"
baseUrl="http://www.cs.qub.ac.uk/itc2007/postenrolcourse/initialdatasets/comp-2007-2-"
ext=".tim"
count=24

if [[ ! -d "${dir}" ]]; then
	mkdir "${dir}"
fi

cd "${dir}"

for i in `seq 1 ${count}`; do
	wget -nv "${baseUrl}${i}${ext}"
done

