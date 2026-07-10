"""ProxyEnv Switch for Windows 11.

Adds, updates, or removes persistent per-user HTTP_PROXY and HTTPS_PROXY
variables. Administrator rights are not required.
"""

from __future__ import annotations

import ctypes
import os
import platform
import sys
import tkinter as tk
from ctypes import wintypes
from tkinter import messagebox, ttk

from proxyenv_core import DEFAULT_HOST, DEFAULT_PORT, VARIABLES, build_proxy_url, validate_port

if platform.system() != "Windows":
    raise SystemExit("ProxyEnv Switch is designed for Windows 11.")

import winreg  # Windows-only standard library module

APP_TITLE = "ProxyEnv Switch"
APP_VERSION = "1.2.0"
ENV_KEY_PATH = r"Environment"


def enable_dpi_awareness() -> None:
    """Enable the best DPI-awareness mode supported by the current Windows build."""
    try:
        user32 = ctypes.WinDLL("user32", use_last_error=True)
        set_context = getattr(user32, "SetProcessDpiAwarenessContext", None)
        if set_context is not None:
            set_context.argtypes = [ctypes.c_void_p]
            set_context.restype = wintypes.BOOL
            # DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 is the pointer value -4.
            if set_context(ctypes.c_void_p(-4)):
                return
    except (AttributeError, OSError):
        pass

    try:
        shcore = ctypes.WinDLL("shcore", use_last_error=True)
        set_awareness = shcore.SetProcessDpiAwareness
        set_awareness.argtypes = [ctypes.c_int]
        set_awareness.restype = ctypes.c_long
        # PROCESS_PER_MONITOR_DPI_AWARE = 2.
        if set_awareness(2) == 0:
            return
    except (AttributeError, OSError):
        pass

    try:
        user32 = ctypes.WinDLL("user32", use_last_error=True)
        user32.SetProcessDPIAware()
    except (AttributeError, OSError):
        pass


def broadcast_environment_change() -> bool:
    """Notify Windows that the current user's environment has changed.

    The registry update remains valid even if one application does not respond to
    the notification. Existing processes still need to be restarted.
    """
    try:
        user32 = ctypes.WinDLL("user32", use_last_error=True)
        send_message_timeout = user32.SendMessageTimeoutW
        send_message_timeout.argtypes = (
            wintypes.HWND,
            wintypes.UINT,
            wintypes.WPARAM,
            wintypes.LPARAM,
            wintypes.UINT,
            wintypes.UINT,
            ctypes.POINTER(ctypes.c_size_t),
        )
        send_message_timeout.restype = wintypes.LPARAM

        environment_text = ctypes.c_wchar_p("Environment")
        result = ctypes.c_size_t()
        sent = send_message_timeout(
            wintypes.HWND(0xFFFF),  # HWND_BROADCAST
            0x001A,  # WM_SETTINGCHANGE
            0,
            ctypes.cast(environment_text, ctypes.c_void_p).value,
            0x0002,  # SMTO_ABORTIFHUNG
            5000,
            ctypes.byref(result),
        )
        return bool(sent)
    except (AttributeError, OSError, TypeError):
        return False


def read_user_variable(name: str) -> str | None:
    """Read one current-user environment value from the registry."""
    try:
        with winreg.OpenKey(
            winreg.HKEY_CURRENT_USER,
            ENV_KEY_PATH,
            0,
            winreg.KEY_QUERY_VALUE,
        ) as key:
            value, _value_type = winreg.QueryValueEx(key, name)
            return str(value)
    except FileNotFoundError:
        return None


def set_user_variable(name: str, value: str) -> None:
    """Create or replace one current-user environment value."""
    with winreg.CreateKeyEx(
        winreg.HKEY_CURRENT_USER,
        ENV_KEY_PATH,
        0,
        winreg.KEY_SET_VALUE,
    ) as key:
        winreg.SetValueEx(key, name, 0, winreg.REG_SZ, value)


def delete_user_variable(name: str) -> bool:
    """Delete one current-user environment value, returning whether it existed."""
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


def snapshot_proxy_variables() -> dict[str, str | None]:
    return {name: read_user_variable(name) for name in VARIABLES}


def restore_proxy_variables(snapshot: dict[str, str | None]) -> None:
    """Best-effort restoration used when a two-value update fails partway through."""
    errors: list[str] = []
    for name, previous_value in snapshot.items():
        try:
            if previous_value is None:
                delete_user_variable(name)
            else:
                set_user_variable(name, previous_value)
        except OSError as exc:
            errors.append(f"{name}: {exc}")

    if errors:
        raise OSError("; ".join(errors))


def update_proxy_variables(proxy_url: str) -> None:
    """Set both proxy values, verify them, and roll back on partial failure."""
    previous = snapshot_proxy_variables()
    try:
        for variable in VARIABLES:
            set_user_variable(variable, proxy_url)

        for variable in VARIABLES:
            if read_user_variable(variable) != proxy_url:
                raise OSError(f"Windows did not preserve the expected {variable} value.")
    except OSError as exc:
        try:
            restore_proxy_variables(previous)
        except OSError as rollback_exc:
            raise OSError(f"{exc}\nRollback also failed: {rollback_exc}") from exc
        raise


def remove_proxy_variables() -> bool:
    """Remove both proxy values, verify removal, and roll back on partial failure."""
    previous = snapshot_proxy_variables()
    existed = any(value is not None for value in previous.values())
    try:
        for variable in VARIABLES:
            delete_user_variable(variable)

        remaining = [name for name in VARIABLES if read_user_variable(name) is not None]
        if remaining:
            raise OSError("Windows did not remove: " + ", ".join(remaining))
    except OSError as exc:
        try:
            restore_proxy_variables(previous)
        except OSError as rollback_exc:
            raise OSError(f"{exc}\nRollback also failed: {rollback_exc}") from exc
        raise
    return existed


class ProxyEnvSwitch(tk.Tk):
    def __init__(self) -> None:
        super().__init__()
        self.title(f"{APP_TITLE} {APP_VERSION}")
        self.geometry("760x520")
        self.minsize(680, 480)
        self.protocol("WM_DELETE_WINDOW", self.destroy)

        self.port_var = tk.StringVar(value=DEFAULT_PORT)
        self.status_var = tk.StringVar()
        self.current_http_var = tk.StringVar()
        self.current_https_var = tk.StringVar()

        self._configure_style()
        self._build_ui()
        self.refresh_status(update_message=True)
        self.after(100, self._center_window)

    def _configure_style(self) -> None:
        style = ttk.Style(self)
        if "vista" in style.theme_names():
            style.theme_use("vista")
        style.configure("Title.TLabel", font=("Segoe UI", 18, "bold"))
        style.configure("Heading.TLabel", font=("Segoe UI", 10, "bold"))
        style.configure("Status.TLabel", font=("Segoe UI", 10))
        style.configure("Note.TLabel", foreground="#555555")
        style.configure("Action.TButton", padding=(12, 9))

    def _build_ui(self) -> None:
        self.columnconfigure(0, weight=1)
        self.rowconfigure(0, weight=1)

        outer = ttk.Frame(self, padding=(28, 24, 28, 20))
        outer.grid(row=0, column=0, sticky="nsew")
        outer.columnconfigure(0, weight=1)
        outer.rowconfigure(5, weight=1)

        title_row = ttk.Frame(outer)
        title_row.grid(row=0, column=0, sticky="ew")
        title_row.columnconfigure(0, weight=1)
        ttk.Label(
            title_row,
            text="Windows Proxy Environment Variables",
            style="Title.TLabel",
        ).grid(row=0, column=0, sticky="w")
        ttk.Label(title_row, text=f"v{APP_VERSION}", style="Note.TLabel").grid(
            row=0, column=1, sticky="e", padx=(12, 0)
        )

        self.subtitle_label = ttk.Label(
            outer,
            text=(
                "Add, update, or remove persistent user-level HTTP_PROXY and "
                "HTTPS_PROXY values. Administrator rights are not required."
            ),
            justify="left",
        )
        self.subtitle_label.grid(row=1, column=0, sticky="ew", pady=(7, 20))

        input_frame = ttk.Frame(outer)
        input_frame.grid(row=2, column=0, sticky="ew")
        input_frame.columnconfigure(1, weight=1)

        ttk.Label(input_frame, text="Proxy address:", style="Heading.TLabel").grid(
            row=0, column=0, sticky="w", padx=(0, 12)
        )

        address_frame = ttk.Frame(input_frame)
        address_frame.grid(row=0, column=1, sticky="w")
        ttk.Label(address_frame, text=f"http://{DEFAULT_HOST}:").grid(
            row=0, column=0, sticky="w"
        )
        validate_digits = (self.register(self._validate_port_entry), "%P")
        self.port_entry = ttk.Entry(
            address_frame,
            textvariable=self.port_var,
            width=10,
            justify="center",
            validate="key",
            validatecommand=validate_digits,
        )
        self.port_entry.grid(row=0, column=1, sticky="w", padx=(4, 0))
        self.port_entry.focus_set()
        self.port_entry.select_range(0, "end")
        self.port_entry.bind("<Return>", lambda _event: self.add_proxy())

        button_frame = ttk.Frame(outer)
        button_frame.grid(row=3, column=0, sticky="ew", pady=(20, 20))
        button_frame.columnconfigure(0, weight=1)
        button_frame.columnconfigure(1, weight=1)

        ttk.Button(
            button_frame,
            text="Add / Update Proxy",
            command=self.add_proxy,
            style="Action.TButton",
        ).grid(row=0, column=0, sticky="ew", padx=(0, 7))

        ttk.Button(
            button_frame,
            text="Remove Proxy",
            command=self.remove_proxy,
            style="Action.TButton",
        ).grid(row=0, column=1, sticky="ew", padx=(7, 0))

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

        ttk.Button(current_frame, text="Refresh", command=self.refresh_status).grid(
            row=2, column=1, sticky="e", pady=(10, 0)
        )

        self.status_label = ttk.Label(
            outer,
            textvariable=self.status_var,
            style="Status.TLabel",
            justify="left",
            anchor="nw",
        )
        self.status_label.grid(row=5, column=0, sticky="nsew", pady=(18, 8))

        self.note_label = ttk.Label(
            outer,
            text=(
                "This changes environment variables only; it does not change the Windows "
                "system proxy. Reopen existing terminals and applications after a change."
            ),
            style="Note.TLabel",
            justify="left",
        )
        self.note_label.grid(row=6, column=0, sticky="ew")

        outer.bind("<Configure>", self._update_wrap_lengths)
        self.bind("<F5>", lambda _event: self.refresh_status())

    @staticmethod
    def _validate_port_entry(proposed: str) -> bool:
        return proposed == "" or (proposed.isdigit() and len(proposed) <= 5)

    def _update_wrap_lengths(self, event: tk.Event) -> None:
        wrap = max(int(event.width) - 8, 320)
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

    def refresh_status(self, update_message: bool = True) -> None:
        http_value = read_user_variable("HTTP_PROXY")
        https_value = read_user_variable("HTTPS_PROXY")
        def display(value: str | None) -> str:
            if value is None:
                return "(not set)"
            if value == "":
                return "(empty value)"
            return value

        self.current_http_var.set(display(http_value))
        self.current_https_var.set(display(https_value))

        if not update_message:
            return
        if http_value is None and https_value is None:
            self.status_var.set("No user-level proxy environment variables are currently configured.")
        elif http_value == https_value and http_value is not None:
            self.status_var.set(f"Both proxy variables are configured as {http_value}.")
        else:
            self.status_var.set(
                "Warning: HTTP_PROXY and HTTPS_PROXY are not set to the same value. "
                "Choose Add / Update Proxy to make them consistent, or Remove Proxy to clear them."
            )

    def add_proxy(self) -> None:
        try:
            port = validate_port(self.port_var.get())
            proxy_url = build_proxy_url(port)
            update_proxy_variables(proxy_url)

            for variable in VARIABLES:
                os.environ[variable] = proxy_url
            notified = broadcast_environment_change()
            self.refresh_status(update_message=False)

            message = (
                "Proxy variables were added successfully. Reopen applications or terminals "
                "that were already running so they inherit the new values."
            )
            if not notified:
                message += " Windows could not notify every open application, but the saved values are valid."
            self.status_var.set(message)
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
                f"Windows could not update both variables safely.\n\n{exc}",
                parent=self,
            )
            self.refresh_status(update_message=True)

    def remove_proxy(self) -> None:
        try:
            existed = remove_proxy_variables()
            for variable in VARIABLES:
                os.environ.pop(variable, None)
            notified = broadcast_environment_change()
            self.refresh_status(update_message=False)

            if existed:
                message = (
                    "Proxy variables were removed successfully. Reopen applications or terminals "
                    "that were already running to clear inherited values."
                )
                if not notified:
                    message += " Windows could not notify every open application, but the values were removed."
                self.status_var.set(message)
                messagebox.showinfo(
                    APP_TITLE,
                    "HTTP_PROXY and HTTPS_PROXY were removed from the current user's environment.",
                    parent=self,
                )
            else:
                self.status_var.set(
                    "No user-level HTTP_PROXY or HTTPS_PROXY values were present. Nothing needed to be removed."
                )
                messagebox.showinfo(
                    APP_TITLE,
                    "The proxy variables were already absent.",
                    parent=self,
                )
        except OSError as exc:
            messagebox.showerror(
                APP_TITLE,
                f"Windows could not remove both variables safely.\n\n{exc}",
                parent=self,
            )
            self.refresh_status(update_message=True)


def main() -> None:
    enable_dpi_awareness()
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
