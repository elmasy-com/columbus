#!/bin/bash

WORKDIR="/mnt/md0/test"

date
echo
echo -e "- full:\t\t$(cat ${WORKDIR}/full/* | wc -l)\t| non unique: $(cat ${WORKDIR}/full/* | uniq -D | wc -l)"
echo -e "- domain:\t$(cat ${WORKDIR}/domain/* | wc -l)\t| non unique: $(cat ${WORKDIR}/domain/* | uniq -D | wc -l)"
echo -e "- sub:\t\t$(cat ${WORKDIR}/sub/* | wc -l)\t| non unique: $(cat ${WORKDIR}/sub/* | uniq -D | wc -l)"
echo -e "- tld:\t\t$(cat ${WORKDIR}/tld | wc -l)\t| non unique: $(cat ${WORKDIR}/tld | uniq -D | wc -l)"
echo
cat "${WORKDIR}/index"