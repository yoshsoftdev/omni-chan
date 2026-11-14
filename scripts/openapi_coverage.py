#!/usr/bin/env python3
"""Report backend routes missing from docs/openapi/v1.yaml (TODO 6.6, S6).

Usage: python3 -I scripts/openapi_coverage.py [--list] [--check]
--check exits 1 when any route is undocumented (CI gate: every route must be in the spec).
"""
import collections
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
ROUTES = ROOT / "services/backend-service/internal/api/v1/routes.go"
SPEC = ROOT / "docs/openapi/v1.yaml"


def backend_routes():
    groups, routes = {}, []
    for line in ROUTES.read_text().splitlines():
        if m := re.search(r'(\w+)\s*:=\s*(\w+)\.Group\("([^"]*)"', line):
            var, parent, path = m.groups()
            groups[var] = groups.get(parent, "") + path
        elif m := re.search(r'(\w+)\.(Get|Post|Put|Patch|Delete)\("([^"]*)"', line):
            var, method, path = m.groups()
            full = (groups.get(var, "") + path).rstrip("/") or "/"
            routes.append((method.upper(), re.sub(r":(\w+)", r"{\1}", full)))
    return routes


def documented_operations():
    """(METHOD, path shape) for every operation in the spec."""
    ops, path = set(), None
    for line in SPEC.read_text().splitlines():
        if m := re.match(r"^  (/[^:]*):\s*$", line):
            path = m.group(1)
        elif path and (m := re.match(r"^    (get|post|put|patch|delete):", line)):
            ops.add((m.group(1).upper(), shape(path)))
        elif line and not line.startswith(" "):
            path = None  # left the paths: section
    return ops


def shape(path):
    return re.sub(r"\{[^}]+\}", "{}", path)


def main():
    documented = documented_operations()
    routes = backend_routes()
    missing = [r for r in routes if (r[0], shape(r[1])) not in documented]
    print(f"OpenAPI coverage: {len(routes) - len(missing)}/{len(routes)} backend operations documented")
    by_group = collections.Counter(path.split("/")[1] for _, path in missing)
    for group, n in by_group.most_common():
        print(f"  /{group}: {n} missing")
    if "--list" in sys.argv:
        for method, path in sorted(missing, key=lambda r: (r[1], r[0])):
            print(f"    {method:6} {path}")
    if "--check" in sys.argv and missing:
        sys.exit(1)


if __name__ == "__main__":
    main()
