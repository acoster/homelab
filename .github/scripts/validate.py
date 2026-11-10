#!/usr/bin/env python3
"""Repo-specific checks for the homelab configs.

Run from the repo root: python .github/scripts/validate.py

Checks:
  1. Every YAML/JSON/TOML file parses, and YAML mappings have no duplicate keys
     (a duplicated compose label or env var silently overrides the first one).
  2. Every *.enc.env file is fully sops-encrypted and covered by a .sops.yaml
     creation rule.
  3. Every .doco-cd.<host>.yaml points at directories and files that exist, and
     every compose stack is deployed by doco-cd (except deliberately manual ones).
  4. `docker compose config` succeeds for every stack, with dummy values for the
     keys in its secrets.enc.env, and no variable is left unset.
"""

from __future__ import annotations

import json
import os
import re
import subprocess
import sys
import tomllib
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[2]

# Stacks that are run by hand or by cron instead of doco-cd.
MANUAL_STACKS = {"autorestic"}

errors: list[str] = []


def error(path: Path | str, msg: str, line: int | None = None) -> None:
    rel = Path(path).relative_to(ROOT) if Path(path).is_absolute() else Path(path)
    loc = f"file={rel}" + (f",line={line}" if line else "")
    errors.append(f"::error {loc}::{msg}")


def tracked_files() -> list[Path]:
    out = subprocess.run(
        ["git", "ls-files", "-z"], cwd=ROOT, check=True, capture_output=True
    ).stdout.decode()
    return [ROOT / p for p in out.split("\0") if p]


# --- 1. Parsing --------------------------------------------------------------


class UniqueKeyLoader(yaml.SafeLoader):
    pass


def _construct_mapping(loader: UniqueKeyLoader, node: yaml.MappingNode, deep=False):
    seen: dict = {}
    for key_node, _ in node.value:
        if key_node.tag == "tag:yaml.org,2002:merge":  # `<<: *anchor`
            continue
        key = loader.construct_object(key_node, deep=deep)
        if key in seen:
            raise yaml.constructor.ConstructorError(
                None, None, f"duplicate key {key!r}", key_node.start_mark
            )
        seen[key] = True
    return loader.construct_mapping(node, deep=deep)


UniqueKeyLoader.add_constructor(
    yaml.resolver.BaseResolver.DEFAULT_MAPPING_TAG, _construct_mapping
)


def check_parsing(files: list[Path]) -> None:
    for f in files:
        if f.is_symlink() or not f.is_file():
            continue
        try:
            if f.suffix in {".yaml", ".yml"}:
                list(yaml.load_all(f.read_text(), Loader=UniqueKeyLoader))
            elif f.suffix == ".json":
                json.loads(f.read_text())
            elif f.suffix == ".toml":
                tomllib.loads(f.read_text())
        except yaml.MarkedYAMLError as e:
            mark = e.problem_mark or e.context_mark
            error(f, f"YAML: {e.problem or e}", mark.line + 1 if mark else None)
        except (yaml.YAMLError, json.JSONDecodeError, tomllib.TOMLDecodeError) as e:
            error(f, f"parse error: {e}")


# --- 2. sops -----------------------------------------------------------------

ENC_VALUE = re.compile(r"^ENC\[AES256_GCM,data:.*,type:\w+\]$")


def check_sops(files: list[Path]) -> None:
    rules = yaml.safe_load((ROOT / ".sops.yaml").read_text())["creation_rules"]
    patterns = [re.compile(r["path_regex"]) for r in rules if "path_regex" in r]

    for f in files:
        if not f.name.endswith(".enc.env"):
            continue
        rel = f.relative_to(ROOT).as_posix()
        if not any(p.search(rel) for p in patterns):
            error(f, "not covered by any .sops.yaml creation_rule")

        meta: set[str] = set()
        for n, raw in enumerate(f.read_text().splitlines(), start=1):
            line = raw.strip()
            if not line or line.startswith("#"):
                continue
            key, sep, value = line.partition("=")
            if not sep:
                error(f, "line is not KEY=VALUE", n)
            elif key.startswith("sops_"):
                meta.add(key)
            elif not ENC_VALUE.match(value):
                error(f, f"{key} is not sops-encrypted", n)
        for required in ("sops_mac", "sops_version", "sops_lastmodified"):
            if required not in meta:
                error(f, f"missing sops metadata ({required}); file is not encrypted")


# --- 3. doco-cd --------------------------------------------------------------


def check_doco_cd() -> set[Path]:
    deployed: set[Path] = set()
    for cfg in sorted(ROOT.glob(".doco-cd.*.yaml")):
        for doc in yaml.safe_load_all(cfg.read_text()):
            if not doc:
                continue
            name = doc.get("name", "?")
            wd = ROOT / doc.get("working_dir", "")
            if not wd.is_dir():
                error(cfg, f"{name}: working_dir {doc.get('working_dir')} does not exist")
                continue
            deployed.add(wd.resolve())
            for key in ("compose_files", "env_files"):
                for item in doc.get(key, []):
                    if not (wd / item).exists():
                        error(cfg, f"{name}: {key} entry {wd.relative_to(ROOT)}/{item} does not exist")
    return deployed


# --- 4. docker compose -------------------------------------------------------

# compose logs in logfmt, so the quotes arrive escaped: The \"FOO\" variable ...
UNSET = re.compile(r'The \\?"(\w+)\\?" variable is not set')


def secret_keys(stack: Path) -> list[str]:
    f = stack / "secrets.enc.env"
    if not f.exists():
        return []
    keys = []
    for line in f.read_text().splitlines():
        key = line.partition("=")[0].strip()
        if key and not key.startswith(("#", "sops_")):
            keys.append(key)
    return keys


def check_compose(stacks: list[Path], deployed: set[Path]) -> None:
    for stack in stacks:
        compose = stack / "compose.yaml"
        if stack.resolve() not in deployed and stack.name not in MANUAL_STACKS:
            error(compose, "stack is not deployed by any .doco-cd.<host>.yaml")

        # Clean environment so runner variables can't mask a missing one.
        env = {"PATH": os.environ["PATH"], "HOME": os.environ.get("HOME", "/tmp")}
        env.update({k: "dummy" for k in secret_keys(stack)})
        res = subprocess.run(
            ["docker", "compose", "-f", "compose.yaml", "config", "-q"],
            cwd=stack, env=env, capture_output=True, text=True,
        )
        if res.returncode != 0:
            error(compose, f"docker compose config failed: {res.stderr.strip()}")
            continue
        for var in sorted(set(UNSET.findall(res.stderr))):
            error(compose, f"${{{var}}} is not set in .env or secrets.enc.env")


def main() -> int:
    files = tracked_files()
    stacks = sorted({f.parent for f in files if f.name == "compose.yaml"})

    check_parsing(files)
    check_sops(files)
    deployed = check_doco_cd()
    check_compose(stacks, deployed)

    for e in errors:
        print(e)
    print(f"Checked {len(stacks)} compose stacks; {len(errors)} error(s).")
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
