#!/bin/bash

# CI files to run at all must has the right suffix *-{,no}values.yaml{,.tpl}
# To avoid creating CI cases that will not run in CT we will check using
# this script if any files added to a folder do not meet the expected suffix

# valid path to files example: "/path/to/ci"
FILES=${1-"charts/redpanda/ci"}

exitCode=0
for file in ${FILES}/*
do
 echo verifying file: ${file}
 if [[ $file == *-values.yaml ]] || [[ $file == *-values.yaml.tpl ]] || [[ $file == *-novalues.yaml.tpl ]] || [[ $file == *-novalues.yaml ]]; then
   continue
 else
   echo "- file does not match neither suffix '-values.yaml' nor -values.yaml.tpl'"
   exitCode=1
 fi 
done

exit $exitCode

