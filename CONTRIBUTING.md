## Contributing

One way to debug during helm development is to diff the kubernetes configuration files that helm generates:

Create and initialize git in a new directory (make sure to be outside an existing source-controlled directory):

```sh
mkdir helm-output
cd helm-output
git init
```

Create a new file `split-output.sh` in the `helm-output` directory:

```
#!/bin/sh
# split-output.sh - split helm output into individual files
csplit -s --suppress-matched -z helm-output.yaml /---/ '{*}'
# remove the first file (it is helm metadata rather than a k8s object)
rm xx00
# loop through each file and rename according to to k8s name and source
# ex. <name>-<source-file>, or redpanda-statefulset.yaml
for file in xx*; do
  NEWNAME=`grep -Po "(?<=^\ \ name:\ ).*" $file | sed 's/"//g' | xargs`-`head -1 "$file" | cut -d '/' -f 3 | sed 's/\.yaml//'`.yaml
  mv "$file" $NEWNAME
done
```

Create another new file `get-redpanda-config.sh` in the same directory:

```
#!/bin/sh
# get-redpanda-config.sh - retrieves Redpanda config from a running node
# first argument is the node name
# second argument is the application.cc line number for that node
# third argument is the output file
# ex: ./get-redpanda-logs.sh redpanda-0 327 ../helm-output/redpanda-0-config.txt
kubectl logs -n helm-test $1 | grep application.cc:$2 | awk -F' - ' '{ $1=""; $2=""; print}' > $3
```

Then do a dry run of the helm install before you make changes to the code:

```sh
helm -n helm-test install redpanda redpanda --dry-run > ../helm-output/helm-output.yaml
```


