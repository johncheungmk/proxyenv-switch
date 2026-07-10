"""Unit tests for platform-independent ProxyEnv Switch helpers."""

from __future__ import annotations

import sys
import unittest
from pathlib import Path

SRC = Path(__file__).resolve().parents[1] / "src"
sys.path.insert(0, str(SRC))

from proxyenv_core import build_proxy_url, validate_port  # noqa: E402


class ValidatePortTests(unittest.TestCase):
    def test_valid_ports(self) -> None:
        self.assertEqual(validate_port("1"), 1)
        self.assertEqual(validate_port(" 60505 "), 60505)
        self.assertEqual(validate_port("65535"), 65535)

    def test_invalid_ports(self) -> None:
        for value in ("", "0", "65536", "abc", "60.5", "-1", "123456"):
            with self.subTest(value=value):
                with self.assertRaises(ValueError):
                    validate_port(value)


class BuildProxyURLTests(unittest.TestCase):
    def test_default_host(self) -> None:
        self.assertEqual(build_proxy_url(60505), "http://127.0.0.1:60505")

    def test_invalid_port(self) -> None:
        with self.assertRaises(ValueError):
            build_proxy_url(0)


if __name__ == "__main__":
    unittest.main()
