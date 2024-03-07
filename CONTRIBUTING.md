# Contributing

## Development Environment

The development environment is managed by [`nix`](https://nixos.org). If
installing nix is not an option, you can ensure all the tools listed in
[`flake.nix`](./flake.nix) are installed and available in `$PATH`.

To install nix, either follow the [official guide](https://nixos.org/download) or [zero-to-nix's quick start](https://zero-to-nix.com/start).

Next, you'll want to enable the experimental features `flakes` and
`nix-command` to avoid typing `--extra-experimental-features nix-command
--extra-experimental-features flakes` all the time.

```bash
# Or open the file in an editor and paste in this line.
# If you're using nix to manage your nix install, you'll have to find your own path :)
echo 'experimental-features = nix-command flakes' >> /etc/nix/nix.conf
```

Now you're ready to go!

```sh
nix develop # Enter a development shell.
task shell # Or enter through task
nix develop -c fish  # Enter a development shell using fish/zsh
nix develop -c zsh  # Enter a development shell using fish/zsh
```

## Contributing

One way to debug during helm development is to diff the kubernetes configuration files that helm generates:

### Setup

Create and initialize git in a new directory (make sure to be outside an existing source-controlled directory):

```sh
mkdir helm-output
cd helm-output
git init
```

#### Create scripts

Create a new file `split-output.sh` in the `helm-output` directory:

```
#!/bin/bash
# split-output.sh - split helm output into individual files
if (($# != 2)); then
  echo "Requires the helm directory and namespace as arguments"
  echo "./split-output.sh <helm directory> <namespace>"
  echo "ex: ./split-output.sh helm-charts/redpanda helm-test"
  exit 1
fi
helm -n $2 install redpanda $1 --dry-run > input.yaml
csplit -s --suppress-matched -z input.yaml /---/ '{*}'
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
#!/bin/bash
# get-redpanda-config.sh - retrieves Redpanda config from a running node
# This node's configuration is printed out at startup, and a file responsible for this is application.cc
# We need to find the relevant message in the log to then filter all subsequent messages by
if (($# != 1)); then
  echo "Requires namespace as the first argument"
  echo "./get-redpanda-config.sh <namespace>"
  echo "ex: ./get-redpanda-config.sh helm-test"
  exit 1
fi
COUNT=$(kubectl -n $1 get sts | sed '2q;d' | awk '{print $2}' | cut -c1-1)
for (( i=0; i<$COUNT; i++)); do
  RELEVANTLINE=$(kubectl logs -n $1 redpanda-$i | grep application.cc | sed '7q;d' | awk '{print $8}')
  kubectl logs -n $1 redpanda-$i | grep $RELEVANTLINE | awk -F' - ' '{ $1=""; $2=""; print}' > ~/projects/redpanda/helm-output/redpanda-$i-config.txt
done
```

Make both of these files executable:

```sh
chmod +x split-output.sh && chmod +x get-redpanda-config.sh
```

#### Run scripts

Collect the initial logs and yaml files from a running Redpanda cluster. The namespace `helm-test` is used below... replace with whatever namespace you used for your redpanda install. This will be the baseline of your output files, so you will be able to compare any differences based on future changes to the helm chart:

```sh
./get-redpanda-config.sh helm-test
./split-output.sh ../helm-charts/redpanda helm-test
```

You will have many yaml files along with a `redpanda-0-config.txt` file and your two scripts. Commit all these files to the repo:

```sh
git add . && git commit -m 'init commit'
```

### Iterate

Now you are ready to make changes to the helm chart and eventually restart your cluster. Once this is done, re-run the `get-redpanda-config.sh` and `split-output.sh` scripts to extract the same files again, and use git to diff the results.

#### Restoring output directory

One you are done comparing the differences, you will likely want to get back to the initial state for these log files so you can compare against subsequent runs. Do this with the following command:


```sh
git restore -- . && git clean -df
```

But be careful! The above command will throw away all changes in whatever git repo you run it against, and you will be back to the most recent commit.

