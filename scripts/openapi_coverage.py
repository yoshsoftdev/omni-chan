#!/usr/bin/env python3
"""Report backend routes missing from docs/openapi/v1.yaml (TODO 6.6, S6).

Usage: python3 -I scripts/openapi_coverage.py [--list]
Exits 0; it is a progress report, not a gate (the spec is incomplete today).
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


def main():
    shape = lambda p: re.sub(r"\{[^}]+\}", "{}", p)
    documented = {shape(p) for p in re.findall(r"^  (/[^:]*):\s*$", SPEC.read_text(), re.M)}
    routes = backend_routes()
    missing = [r for r in routes if shape(r[1]) not in documented]
    print(f"OpenAPI coverage: {len(routes) - len(missing)}/{len(routes)} backend routes have a documented path")
    by_group = collections.Counter(path.split("/")[1] for _, path in missing)
    for group, n in by_group.most_common():
        print(f"  /{group}: {n} missing")
    if "--list" in sys.argv:
        for method, path in sorted(missing, key=lambda r: (r[1], r[0])):
            print(f"    {method:6} {path}")


if __name__ == "__main__":
    main()
