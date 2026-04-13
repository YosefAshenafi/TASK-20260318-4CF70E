#!/usr/bin/env python3
"""
Static conformance checks against docs/design.md (authoritative product design).
Fails the process (exit 1) if required route areas or RBAC seeds are missing.
"""
from __future__ import annotations

import re
import sys
from pathlib import Path


def repo_root() -> Path:
    return Path(__file__).resolve().parents[1]


def design_route_areas(root: Path) -> list[str]:
    design = (root.parent / "docs" / "design.md").read_text(encoding="utf-8")
    m = re.search(r"### 5\.2 Route Areas\s*\n+(.*?)(?=\n### )", design, re.DOTALL)
    if not m:
        print("[design] Missing '### 5.2 Route Areas' in docs/design.md", file=sys.stderr)
        sys.exit(1)
    block = m.group(1)
    paths: list[str] = []
    for line in block.splitlines():
        mm = re.match(r"\* `(/[^`]+)`", line.strip())
        if mm:
            paths.append(mm.group(1))
    if not paths:
        print("[design] No route bullets found under 5.2", file=sys.stderr)
        sys.exit(1)
    return paths


def assert_routes_in_router(root: Path, routes: list[str]) -> None:
    router = (root / "apps" / "web" / "src" / "router" / "index.ts").read_text(encoding="utf-8")
    for r in routes:
        tail = r.lstrip("/")
        if tail not in router and f"'/{tail}'" not in router and f'"/{tail}"' not in router:
            # /login is 'login' path
            seg = tail.split("/")[0]
            if seg not in router:
                print(
                    f"[design] Route area {r!r} not found in src/router/index.ts",
                    file=sys.stderr,
                )
                sys.exit(1)


def assert_menu_covers_dashboard(root: Path, routes: list[str]) -> None:
    layout = (root / "apps" / "web" / "src" / "layouts" / "AppLayout.vue").read_text(encoding="utf-8")
    for r in routes:
        if r in ("/login",):
            continue
        tail = r if r.startswith("/") else "/" + r
        if tail not in layout:
            print(
                f"[design] Route {tail!r} should appear in AppLayout menuItems (§5.3 role-aware nav)",
                file=sys.stderr,
            )
            sys.exit(1)


def assert_theme_layer(root: Path) -> None:
    theme = root / "apps" / "web" / "src" / "theme" / "element-theme.css"
    if not theme.is_file():
        print("[design] Missing apps/web/src/theme/element-theme.css (§5.3 custom theme)", file=sys.stderr)
        sys.exit(1)
    text = theme.read_text(encoding="utf-8")
    if "--el-color-primary" not in text:
        print("[design] element-theme.css should define --el-color-primary (teal palette §5.3)", file=sys.stderr)
        sys.exit(1)


def assert_primary_roles_seed(root: Path) -> None:
    seed = root / "infra" / "db" / "migrations" / "000013_primary_roles_seed.up.sql"
    if not seed.is_file():
        print("[design] Missing 000013_primary_roles_seed migration (§8.1.1)", file=sys.stderr)
        sys.exit(1)
    text = seed.read_text(encoding="utf-8")
    required = (
        "business_specialist",
        "compliance_administrator",
        "recruitment_specialist",
        "system_admin",
    )
    for slug in required:
        if slug not in text:
            print(f"[design] Primary role slug {slug!r} missing from 000013 (§8.1.1)", file=sys.stderr)
            sys.exit(1)


def assert_api_surface(root: Path) -> None:
    server = (root / "apps" / "api" / "internal" / "httpserver" / "server.go").read_text(encoding="utf-8")
    required_snippets = (
        "/api/v1",
        "POST(\"/auth/login\"",
        "SessionAuth",
        "AccessContext",
        "RequirePermission",
        "/recruitment/candidates",
        "/compliance/qualifications",
        "/cases",
        "/files",
        "/audit/logs",
        "/users",
    )
    for snip in required_snippets:
        if snip not in server:
            print(f"[design] server.go missing expected wiring {snip!r} (§6, domain modules)", file=sys.stderr)
            sys.exit(1)


def main() -> int:
    root = repo_root()
    routes = design_route_areas(root)
    print(f"[design] design.md §5.2 route areas ({len(routes)}): {', '.join(routes)}")
    assert_routes_in_router(root, routes)
    assert_menu_covers_dashboard(root, routes)
    assert_theme_layer(root)
    assert_primary_roles_seed(root)
    assert_api_surface(root)
    print("[design] Conformance OK (routes, nav, theme, primary roles, API wiring).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
