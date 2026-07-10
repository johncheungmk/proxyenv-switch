"""Platform-independent helpers for ProxyEnv Switch."""

from __future__ import annotations

import re

VARIABLES = ("HTTP_PROXY", "HTTPS_PROXY")
DEFAULT_HOST = "127.0.0.1"
DEFAULT_PORT = "60505"


def validate_port(raw_port: str) -> int:
    """Validate a TCP port supplied by the user."""
    value = raw_port.strip()
    if not re.fullmatch(r"\d{1,5}", value):
        raise ValueError("Enter a numeric port from 1 to 65535.")

    port = int(value)
    if not 1 <= port <= 65535:
        raise ValueError("The port must be from 1 to 65535.")
    return port


def build_proxy_url(port: int, host: str = DEFAULT_HOST) -> str:
    """Return the HTTP proxy URL stored in the environment variables."""
    if not 1 <= port <= 65535:
        raise ValueError("The port must be from 1 to 65535.")
    return f"http://{host}:{port}"
