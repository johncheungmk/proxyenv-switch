"""ProxyEnv Switch for Windows 11.

Adds, updates, or removes persistent per-user HTTP_PROXY and HTTPS_PROXY
variables. Administrator rights are not required.
"""

from __future__ import annotations

import ctypes
import os
import platform
import re
import sys
import tkinter as tk
from pathlib import Path
from tkinter import messagebox, ttk

if platform.system() != "Windows":
    raise SystemExit("ProxyEnv Switch is designed for Windows 11.")

import winreg  # Windows-only standard library module

APP_TITLE = "ProxyEnv Switch"
APP_VERSION = "1.1.0"
ENV_KEY_PATH = r"Environment"
VARIABLES = ("HTTP_PROXY", "HTTPS_PROXY")
DEFAULT_PORT = "60505"


def broadcast_environment_change() -> None:
    """Notify Windows that the current user's environment has changed."""
    hwnd_broadcast = 0xFFFF
    wm_settingchange = 0x001A
    smto_abortifhung = 0x0002
    result = ctypes.c_ulong()
    ctypes.windll.user32.SendMessageTimeoutW(
        hwnd_broadcast,
        wm_settingchange,
        0,
        "Environment",
        smto_abortifhung,
        5000,
        ctypes.byref(result),
    )


def read_user_variable(name: str) -> str | None:
    try:
        with winreg.OpenKey(winreg.HKEY_CURRENT_USER, ENV_KEY_PATH) as key:
            value, _ = winreg.QueryValueEx(key, name)
            return str(value)
    except FileNotFoundError:
        return None


def set_user_variable(name: str, value: str) -> None:
    with winreg.CreateKeyEx(
        winreg.HKEY_CURRENT_USER,
        ENV_KEY_PATH,
        0,
        winreg.KEY_SET_VALUE,
    ) as key:
        winreg.SetValueEx(key, name, 0, winreg.REG_SZ, value)


def delete_user_variable(name: str) -> bool:
    try:
        with winreg.OpenKey(
            winreg.HKEY_CURRENT_USER,
            ENV_KEY_PATH,
            0,
            winreg.KEY_SET_VALUE,
        ) as key:
            winreg.DeleteValue(key, name)
        return True
    except FileNotFoundError:
        return False


def validate_port(raw_port: str) -> int:
    value = raw_port.strip()
    if not re.fullmatch(r"\d{1,5}", value):
        raise ValueError("Enter a numeric port from 1 to 65535.")
    port = int(value)
    if not 1 <= port <= 65535:
        raise ValueError("The port must be from 1 to 65535.")
    return port


class ProxyEnvSwitch(tk.Tk):
    def __init__(self) -> None:
        super().__init__()
        self.title(APP_TITLE)
        self.geometry("720x480")
        self.minsize(650, 440)
        self.protocol("WM_DELETE_WINDOW", self.destroy)

        self.port_var = tk.StringVar(value=DEFAULT_PORT)
        self.status_var = tk.StringVar(value="Enter a port, then choose an action.")
        self.current_http_var = tk.StringVar()
        self.current_https_var = tk.StringVar()

        self._configure_style()
        self._build_ui()
        self.refresh_status()
        self.after(100, self._center_window)

    def _configure_style(self) -> None:
        style = ttk.Style(self)
        if "vista" in style.theme_names():
            style.theme_use("vista")
        style.configure("Title.TLabel", font=("Segoe UI", 17, "bold"))
        style.configure("Heading.TLabel", font=("Segoe UI", 10, "bold"))
        style.configure("Status.TLabel", font=("Segoe UI", 10))
        style.configure("Note.TLabel", foreground="#555555")
        style.configure("Action.TButton", padding=(12, 8))

    def _build_ui(self) -> None:
        self.columnconfigure(0, weight=1)
        self.rowconfigure(0, weight=1)

        outer = ttk.Frame(self, padding=(26, 22, 26, 20))
        outer.grid(row=0, column=0, sticky="nsew")
        outer.columnconfigure(0, weight=1)
        outer.rowconfigure(5, weight=1)

        ttk.Label(
            outer,
            text="Windows Proxy Environment Variables",
            style="Title.TLabel",
        ).grid(row=0, column=0, sticky="w")

        self.subtitle_label = ttk.Label(
            outer,
            text=(
                "Add, update, or remove persistent user-level HTTP_PROXY and "
                "HTTPS_PROXY values. Administrator rights are not required."
            ),
            justify="left",
        )
        self.subtitle_label.grid(row=1, column=0, sticky="ew", pady=(6, 18))

        input_frame = ttk.Frame(outer)
        input_frame.grid(row=2, column=0, sticky="ew")
        input_frame.columnconfigure(1, weight=1)

        ttk.Label(input_frame, text="Proxy address:", style="Heading.TLabel").grid(
            row=0, column=0, sticky="w", padx=(0, 12)
        )

        address_frame = ttk.Frame(input_frame)
        address_frame.grid(row=0, column=1, sticky="w")
        ttk.Label(address_frame, text="http://127.0.0.1:").grid(row=0, column=0, sticky="w")
        self.port_entry = ttk.Entry(
            address_frame,
            textvariable=self.port_var,
            width=10,
            justify="center",
        )
        self.port_entry.grid(row=0, column=1, sticky="w", padx=(4, 0))
        self.port_entry.focus_set()
        self.port_entry.select_range(0, "end")
        self.port_entry.bind("<Return>", lambda _event: self.add_proxy())

        button_frame = ttk.Frame(outer)
        button_frame.grid(row=3, column=0, sticky="ew", pady=(18, 18))
        button_frame.columnconfigure(0, weight=1)
        button_frame.columnconfigure(1, weight=1)

        ttk.Button(
            button_frame,
            text="Add / Update Proxy",
            command=self.add_proxy,
            style="Action.TButton",
        ).grid(row=0, column=0, sticky="ew", padx=(0, 6))

        ttk.Button(
            button_frame,
            text="Remove Proxy",
            command=self.remove_proxy,
            style="Action.TButton",
        ).grid(row=0, column=1, sticky="ew", padx=(6, 0))

        current_frame = ttk.LabelFrame(outer, text="Current user values", padding=(14, 12))
        current_frame.grid(row=4, column=0, sticky="ew")
        current_frame.columnconfigure(1, weight=1)

        ttk.Label(current_frame, text="HTTP_PROXY:", style="Heading.TLabel").grid(
            row=0, column=0, sticky="w", padx=(0, 10), pady=(0, 8)
        )
        ttk.Entry(
            current_frame,
            textvariable=self.current_http_var,
            state="readonly",
        ).grid(row=0, column=1, sticky="ew", pady=(0, 8))

        ttk.Label(current_frame, text="HTTPS_PROXY:", style="Heading.TLabel").grid(
            row=1, column=0, sticky="w", padx=(0, 10)
        )
        ttk.Entry(
            current_frame,
            textvariable=self.current_https_var,
            state="readonly",
        ).grid(row=1, column=1, sticky="ew")

        self.status_label = ttk.Label(
            outer,
            textvariable=self.status_var,
            style="Status.TLabel",
            justify="left",
            anchor="nw",
        )
        self.status_label.grid(row=5, column=0, sticky="nsew", pady=(16, 8))

        self.note_label = ttk.Label(
            outer,
            text=(
                "Changes apply to the current Windows user only. Reopen existing "
                "terminals and applications after making a change."
            ),
            style="Note.TLabel",
            justify="left",
        )
        self.note_label.grid(row=6, column=0, sticky="ew")

        outer.bind("<Configure>", self._update_wrap_lengths)

    def _update_wrap_lengths(self, event: tk.Event) -> None:
        wrap = max(int(event.width) - 8, 300)
        self.subtitle_label.configure(wraplength=wrap)
        self.status_label.configure(wraplength=wrap)
        self.note_label.configure(wraplength=wrap)

    def _center_window(self) -> None:
        self.update_idletasks()
        width = self.winfo_width()
        height = self.winfo_height()
        x = max((self.winfo_screenwidth() - width) // 2, 0)
        y = max((self.winfo_screenheight() - height) // 2, 0)
        self.geometry(f"{width}x{height}+{x}+{y}")

    def refresh_status(self) -> None:
        self.current_http_var.set(read_user_variable("HTTP_PROXY") or "(not set)")
        self.current_https_var.set(read_user_variable("HTTPS_PROXY") or "(not set)")

    def add_proxy(self) -> None:
        try:
            port = validate_port(self.port_var.get())
            proxy_url = f"http://127.0.0.1:{port}"
            for variable in VARIABLES:
                set_user_variable(variable, proxy_url)
                os.environ[variable] = proxy_url
            broadcast_environment_change()
            self.refresh_status()
            self.status_var.set(
                "Proxy variables were added successfully. Reopen applications or "
                "terminals that were already running so they inherit the new values."
            )
            messagebox.showinfo(
                APP_TITLE,
                f"HTTP_PROXY and HTTPS_PROXY were set to:\n\n{proxy_url}",
                parent=self,
            )
        except ValueError as exc:
            messagebox.showwarning(APP_TITLE, str(exc), parent=self)
        except OSError as exc:
            messagebox.showerror(
                APP_TITLE,
                f"Windows could not update the variables.\n\n{exc}",
                parent=self,
            )

    def remove_proxy(self) -> None:
        try:
            removed = [delete_user_variable(variable) for variable in VARIABLES]
            for variable in VARIABLES:
                os.environ.pop(variable, None)
            broadcast_environment_change()
            self.refresh_status()

            if any(removed):
                self.status_var.set(
                    "Proxy variables were removed successfully. Reopen applications "
                    "or terminals that were already running to clear inherited values."
                )
                messagebox.showinfo(
                    APP_TITLE,
                    "HTTP_PROXY and HTTPS_PROXY were removed from the current user's environment.",
                    parent=self,
                )
            else:
                self.status_var.set(
                    "No user-level HTTP_PROXY or HTTPS_PROXY values were present. "
                    "Nothing needed to be removed."
                )
                messagebox.showinfo(
                    APP_TITLE,
                    "The proxy variables were already absent.",
                    parent=self,
                )
        except OSError as exc:
            messagebox.showerror(
                APP_TITLE,
                f"Windows could not remove the variables.\n\n{exc}",
                parent=self,
            )


def main() -> None:
    try:
        app = ProxyEnvSwitch()
        app.mainloop()
    except Exception as exc:  # Last-resort GUI error for packaged builds.
        try:
            ctypes.windll.user32.MessageBoxW(0, str(exc), APP_TITLE, 0x10)
        finally:
            sys.exit(1)


if __name__ == "__main__":
    main()
