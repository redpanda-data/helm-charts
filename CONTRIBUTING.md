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

## Making Changes

User facing changes should be documented in [the CHANGELOG][./CHANGELOG.md]
under the chart appropriate subheading. Changes are grouped into "Added,
Changed, Fixed, and Removed". Breaking changes[^1] and deprecations should be
prefixed with the "BREAKING" and "DEPRECATED", respectively.

[^1]: A change is considered breaking if the `values.yaml`, sans deprecated fields,
from the previous release cannot be used in a `helm upgrade` command without
modifications.
