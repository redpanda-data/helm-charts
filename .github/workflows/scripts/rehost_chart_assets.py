#!/usr/bin/env python3
"""Re-host chart packages referenced by index.yaml as releases of this repo.

Some entries in the Helm repository index point at chart packages that are
attached to GitHub releases of another repository. This script copies each
such package to a release of the target repository (same release/tag name,
same file) and rewrites the index entry to point at the new location. The
package bytes do not change, so the digest recorded in the index stays valid
and is verified along the way.

Only `urls` lines are touched; the rest of index.yaml is preserved byte for
byte so the resulting diff is reviewable line by line.

Dry run is the default: packages are downloaded and verified and the plan is
printed, but nothing is created or uploaded and the index is only written if
--out names a file other than the input. With --execute the index is
rewritten in place unless --out is given.

A release is created when it does not exist yet, and a package is only
uploaded when the release lacks it or holds a different digest, so the script
can be re-run after a failure and a complete run is a no-op. New release tags
point at the pages branch rather than main, so release.yaml, which diffs
charts/ against the newest tag reachable from main, never sees them.

Usage:
  rehost_chart_assets.py --index index.yaml [--execute] [--out index.yaml]
      [--source-repo redpanda-data/redpanda-operator]
      [--target-repo redpanda-data/helm-charts] [--workdir DIR]
      [--target-branch gh-pages]

Requires: gh; --execute needs write access to the target repository.
"""
import argparse
import hashlib
import json
import os
import re
import subprocess
import sys
import tempfile
from collections import Counter


def sh(*args, check=True, capture=True):
    return subprocess.run(args, check=check, text=True,
                          capture_output=capture)


def sha256(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()


def release_asset_digest(repo, tag, name):
    """Return (release exists, digest of asset `name` or None)."""
    r = sh("gh", "api", f"repos/{repo}/releases/tags/{tag}", check=False)
    if r.returncode != 0:
        return False, None
    for asset in json.loads(r.stdout)["assets"]:
        if asset["name"] == name:
            return True, asset.get("digest")
    return True, None


def main():
    ap = argparse.ArgumentParser(description=__doc__.split("\n\n")[0])
    ap.add_argument("--index", required=True)
    ap.add_argument("--out", help="where to write the rewritten index "
                                  "(default: in place with --execute, "
                                  "not written otherwise)")
    ap.add_argument("--source-repo", default="redpanda-data/redpanda-operator")
    ap.add_argument("--target-repo", default="redpanda-data/helm-charts")
    ap.add_argument("--target-branch", default="gh-pages",
                    help="branch whose head new release tags point at; "
                         "keep it off the default branch, see above")
    ap.add_argument("--workdir", help="where packages are downloaded "
                                      "(default: a temporary directory)")
    ap.add_argument("--execute", action="store_true",
                    help="create releases and upload packages; "
                         "without it only download, verify and report")
    args = ap.parse_args()
    if (args.out and not args.execute
            and os.path.realpath(args.out) == os.path.realpath(args.index)):
        sys.exit("dry run: --out must not be the input file, its urls would "
                 "point at releases that do not exist")

    workdir = args.workdir or tempfile.mkdtemp(prefix="chart-rehost-")
    os.makedirs(workdir, exist_ok=True)

    src_prefix = f"https://github.com/{args.source_repo}/releases/download/"
    dst_prefix = f"https://github.com/{args.target_repo}/releases/download/"
    url_re = re.compile(r"^(\s*- )" + re.escape(src_prefix)
                        + r"(?P<tag>.+)/(?P<file>[^/]+\.tgz)\s*$")
    chart_re = re.compile(r"^  (?P<chart>[A-Za-z0-9._-]+):\s*$")
    digest_re = re.compile(r"^\s{4}digest:\s*(?P<digest>[0-9a-f]{64})\s*$")

    with open(args.index) as f:
        lines = f.read().split("\n")

    chart = None
    digest = None
    plan = []  # (line index, chart, tag, file, digest)
    for i, line in enumerate(lines):
        m = chart_re.match(line)
        if m:
            chart = m.group("chart")
            digest = None
            continue
        if line.startswith("  - apiVersion:"):
            digest = None
            continue
        m = digest_re.match(line)
        if m:
            digest = m.group("digest")
            continue
        m = url_re.match(line)
        if m:
            if digest is None:
                sys.exit(f"line {i + 1}: url without a preceding digest")
            plan.append((i, chart, m.group("tag"), m.group("file"), digest))

    per_chart = Counter(c for _, c, _, _, _ in plan)
    print(f"{len(plan)} package urls point at {args.source_repo}: "
          + ", ".join(f"{c}={n}" for c, n in sorted(per_chart.items())))
    if not plan:
        print("nothing to do")
        return

    created = uploaded = skipped = 0
    for i, chart_name, tag, fname, want in plan:
        dest = os.path.join(workdir, tag.replace("/", "_") + "_" + fname)
        if not os.path.exists(dest):
            ddir = os.path.join(workdir, tag.replace("/", "_"))
            os.makedirs(ddir, exist_ok=True)
            sh("gh", "release", "download", tag, "--repo", args.source_repo,
               "--pattern", fname, "--dir", ddir, "--clobber")
            os.replace(os.path.join(ddir, fname), dest)
        got = sha256(dest)
        if got != want:
            sys.exit(f"{tag}/{fname}: digest mismatch, index has {want}, "
                     f"file is {got}")

        if args.execute:
            exists, have = release_asset_digest(args.target_repo, tag, fname)
            if not exists:
                cmd = ["gh", "release", "create", tag, "--repo",
                       args.target_repo, "--target", args.target_branch,
                       "--title", tag, "--notes",
                       f"Chart package {fname}", "--latest=false"]
                if re.search(r"-(alpha|beta|rc)", tag):
                    cmd.append("--prerelease")
                sh(*cmd)
                created += 1
            if have == f"sha256:{want}":
                skipped += 1
                print(f"have {tag}/{fname}")
            else:
                # The uploaded file must keep its original name: the index
                # url ends with it and Helm derives nothing else from the
                # release. --clobber only matters when the digest differs.
                up = os.path.join(workdir, fname)
                os.replace(dest, up)
                sh("gh", "release", "upload", tag, up, "--repo",
                   args.target_repo, "--clobber")
                os.replace(up, dest)
                uploaded += 1
                print(f"ok   {tag}/{fname}")
        else:
            print(f"plan {tag}/{fname} -> {dst_prefix}{tag}/{fname}")

        lines[i] = lines[i].replace(src_prefix, dst_prefix, 1)

    out = args.out or (args.index if args.execute else None)
    if out:
        with open(out, "w") as f:
            f.write("\n".join(lines))

    mode = "executed" if args.execute else "dry run"
    print(f"{mode}: {len(plan)} packages verified, {created} releases "
          f"created, {skipped} assets already in place, {uploaded} uploads, "
          f"index {'written to ' + out if out else 'not written'}")


if __name__ == "__main__":
    main()
