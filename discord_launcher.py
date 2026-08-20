import os
import sys
import json
import time
import winreg
import subprocess
import tkinter as tk
from tkinter import messagebox, filedialog
from PIL import Image, ImageTk

def get_config_path():
    if getattr(sys, 'frozen', False):
        base_dir = os.path.dirname(sys.executable)
        exe_name = os.path.splitext(os.path.basename(sys.executable))[0]
        return os.path.join(base_dir, f"{exe_name}.config")
    else:
        base_dir = os.path.dirname(os.path.abspath(__file__))
        return os.path.join(base_dir, "DiscordProxyLauncher.config")

CONFIG_FILE = get_config_path()

DEFAULT_CONFIG = {
    "proxy_host": "",
    "proxy_port": "",
    "force_webrtc": True,
    "custom_discord_path": ""
}

# Discord Color Palette
CLR_BG_MAIN     = "#313338"  # Background principal
CLR_BG_CARD     = "#2B2D31"  # Cards / containers
CLR_BG_INPUT    = "#1E1F22"  # Inputs
CLR_BORDER      = "#3F4147"  # Bordas sutis
CLR_BLURPLE     = "#5865F2"  # Discord Blurple
CLR_BLURPLE_HOV = "#4752C4"  # Blurple Hover
CLR_BTN_SEC     = "#4E5058"  # Botão secundário cinza
CLR_BTN_SEC_HOV = "#62656B"  # Botão secundário Hover
CLR_TEXT_WHITE  = "#FFFFFF"  # Branco
CLR_TEXT_NORM   = "#DBDEE1"  # Texto padrão claro
CLR_TEXT_MUTED  = "#949BA4"  # Texto secundário/muted

def load_config():
    if os.path.exists(CONFIG_FILE):
        try:
            with open(CONFIG_FILE, 'r', encoding='utf-8') as f:
                cfg = json.load(f)
                return {**DEFAULT_CONFIG, **cfg}
        except Exception:
            pass
    return DEFAULT_CONFIG.copy()

def save_config(cfg):
    try:
        with open(CONFIG_FILE, 'w', encoding='utf-8') as f:
            json.dump(cfg, f, indent=2, ensure_ascii=False)
        return True
    except Exception:
        return False

def is_discord_running():
    try:
        output = subprocess.check_output('tasklist /FI "IMAGENAME eq Discord.exe" /NH', shell=True, text=True)
        return "Discord.exe" in output
    except Exception:
        return False

def kill_discord():
    try:
        subprocess.run('taskkill /F /IM Discord.exe /T', shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        subprocess.run('taskkill /F /IM DiscordCanary.exe /T', shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        subprocess.run('taskkill /F /IM DiscordPTB.exe /T', shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        time.sleep(0.35)
    except Exception:
        pass

def find_discord_path():
    try:
        with winreg.OpenKey(winreg.HKEY_CURRENT_USER, r"Software\Classes\discord\shell\open\command") as key:
            cmd, _ = winreg.QueryValueEx(key, "")
            if cmd:
                parts = cmd.strip().split('"')
                raw_path = parts[1] if len(parts) > 1 else cmd.split()[0]
                if os.path.basename(raw_path).lower() == 'discord.exe' and os.path.exists(raw_path):
                    return raw_path
                dir_name = os.path.dirname(raw_path)
                found = find_discord_in_dir(dir_name)
                if found:
                    return found
                if os.path.exists(raw_path):
                    return raw_path
    except Exception:
        pass

    try:
        with winreg.OpenKey(winreg.HKEY_CURRENT_USER, r"Software\Microsoft\Windows\CurrentVersion\Uninstall\Discord") as key:
            icon, _ = winreg.QueryValueEx(key, "DisplayIcon")
            if icon:
                icon_path = icon.strip('"').split(',')[0].strip()
                if os.path.basename(icon_path).lower() == 'discord.exe' and os.path.exists(icon_path):
                    return icon_path
                found = find_discord_in_dir(os.path.dirname(icon_path))
                if found:
                    return found
    except Exception:
        pass

    local_app_data = os.environ.get('LOCALAPPDATA', '')
    if local_app_data:
        for candidate in ['Discord', 'DiscordCanary', 'DiscordPTB']:
            base_dir = os.path.join(local_app_data, candidate)
            found = find_discord_in_dir(base_dir)
            if found:
                return found
            update_exe = os.path.join(base_dir, 'Update.exe')
            if os.path.exists(update_exe):
                return update_exe

    for pf in [os.environ.get('ProgramFiles'), os.environ.get('ProgramFiles(x86)'), r'C:\Program Files', r'C:\Program Files (x86)']:
        if pf:
            target = os.path.join(pf, 'Discord', 'Discord.exe')
            if os.path.exists(target):
                return target

    return ""

def find_discord_in_dir(dir_path):
    if not os.path.isdir(dir_path):
        return ""
    direct = os.path.join(dir_path, 'Discord.exe')
    if os.path.exists(direct):
        return direct
    
    try:
        subdirs = [d for d in os.listdir(dir_path) if d.startswith('app-') and os.path.isdir(os.path.join(dir_path, d))]
        subdirs.sort(reverse=True)
        for sub in subdirs:
            exe = os.path.join(dir_path, sub, 'Discord.exe')
            if os.path.exists(exe):
                return exe
    except Exception:
        pass
    return ""

class DiscordProxyLauncherApp(tk.Tk):
    def __init__(self):
        super().__init__()
        self.title("Discord Proxy Launcher")
        self.geometry("580x515")
        self.resizable(False, False)
        self.configure(bg=CLR_BG_MAIN)
        
        base_dir = os.path.dirname(os.path.abspath(__file__))
        ico_path = os.path.join(base_dir, "icon.ico")
        png_path = os.path.join(base_dir, "icon.png")
        if os.path.exists(ico_path):
            try:
                self.iconbitmap(ico_path)
            except Exception:
                pass
        
        self.header_img = None
        if os.path.exists(png_path):
            try:
                pil_img = Image.open(png_path).resize((40, 40), Image.Resampling.LANCZOS)
                self.header_img = ImageTk.PhotoImage(pil_img)
            except Exception:
                pass

        self.config_data = load_config()
        self.init_ui()
        self.protocol("WM_DELETE_WINDOW", self.on_close)

    def create_card(self, parent, y, height):
        card = tk.Frame(parent, bg=CLR_BG_CARD, highlightbackground=CLR_BORDER, highlightthickness=1)
        card.place(x=20, y=y, width=540, height=height)
        return card

    def init_ui(self):
        # 1. Header
        if self.header_img:
            lbl_icon = tk.Label(self, image=self.header_img, bg=CLR_BG_MAIN)
            lbl_icon.place(x=20, y=14)
            lbl_title = tk.Label(self, text="DISCORD PROXY LAUNCHER", font=("Segoe UI", 15, "bold"), fg=CLR_TEXT_WHITE, bg=CLR_BG_MAIN)
            lbl_title.place(x=68, y=14)
            lbl_sub = tk.Label(self, text="Inicie o Discord com proxy SOCKS5 e proteção contra vazamento WebRTC", font=("Segoe UI", 9), fg=CLR_TEXT_MUTED, bg=CLR_BG_MAIN)
            lbl_sub.place(x=68, y=38)
        else:
            lbl_title = tk.Label(self, text="DISCORD PROXY LAUNCHER", font=("Segoe UI", 15, "bold"), fg=CLR_TEXT_WHITE, bg=CLR_BG_MAIN)
            lbl_title.place(x=20, y=14)
            lbl_sub = tk.Label(self, text="Inicie o Discord com proxy SOCKS5 e proteção contra vazamento WebRTC", font=("Segoe UI", 9), fg=CLR_TEXT_MUTED, bg=CLR_BG_MAIN)
            lbl_sub.place(x=20, y=38)

        # 2. Card 1: Proxy Settings
        card_proxy = self.create_card(self, 68, 112)

        lbl_host = tk.Label(card_proxy, text="ENDEREÇO DO PROXY (IP OU HOST)", font=("Segoe UI", 8, "bold"), fg=CLR_TEXT_MUTED, bg=CLR_BG_CARD)
        lbl_host.place(x=15, y=10)

        lbl_port = tk.Label(card_proxy, text="PORTA", font=("Segoe UI", 8, "bold"), fg=CLR_TEXT_MUTED, bg=CLR_BG_CARD)
        lbl_port.place(x=360, y=10)

        self.var_host = tk.StringVar(value=self.config_data.get("proxy_host", ""))
        self.var_port = tk.StringVar(value=self.config_data.get("proxy_port", ""))

        entry_host = tk.Entry(
            card_proxy,
            textvariable=self.var_host,
            font=("Segoe UI", 10),
            bg=CLR_BG_INPUT,
            fg=CLR_TEXT_NORM,
            insertbackground=CLR_TEXT_WHITE,
            relief=tk.FLAT,
            highlightbackground=CLR_BORDER,
            highlightthickness=1
        )
        entry_host.place(x=15, y=30, width=330, height=30)

        entry_port = tk.Entry(
            card_proxy,
            textvariable=self.var_port,
            font=("Segoe UI", 10),
            bg=CLR_BG_INPUT,
            fg=CLR_TEXT_NORM,
            insertbackground=CLR_TEXT_WHITE,
            relief=tk.FLAT,
            highlightbackground=CLR_BORDER,
            highlightthickness=1
        )
        entry_port.place(x=360, y=30, width=165, height=30)

        lbl_tip = tk.Label(card_proxy, text="Protocolo padrão: socks5:// (ex: 127.0.0.1 porta 1080)", font=("Segoe UI", 8), fg=CLR_TEXT_MUTED, bg=CLR_BG_CARD)
        lbl_tip.place(x=15, y=74)

        # 3. Card 2: WebRTC
        card_webrtc = self.create_card(self, 194, 104)

        self.var_webrtc = tk.BooleanVar(value=self.config_data.get("force_webrtc", True))
        chk_webrtc = tk.Checkbutton(
            card_webrtc,
            text=" Forçar tunelamento WebRTC",
            variable=self.var_webrtc,
            font=("Segoe UI", 10, "bold"),
            fg=CLR_TEXT_WHITE,
            bg=CLR_BG_CARD,
            activebackground=CLR_BG_CARD,
            activeforeground=CLR_TEXT_WHITE,
            selectcolor=CLR_BG_INPUT,
            relief=tk.FLAT
        )
        chk_webrtc.place(x=15, y=10)

        lbl_webrtc_desc = tk.Label(
            card_webrtc,
            text="Bloqueia vazamento de IP real desativando conexões UDP diretas (--force-webrtc-ip-handling-policy=disable_non_proxied_udp).",
            font=("Segoe UI", 8),
            fg=CLR_TEXT_MUTED,
            bg=CLR_BG_CARD,
            wraplength=500,
            justify=tk.LEFT
        )
        lbl_webrtc_desc.place(x=38, y=38)

        # 4. Card 3: Discord Location
        card_path = self.create_card(self, 312, 112)

        lbl_path_title = tk.Label(card_path, text="LOCALIZAÇÃO DO DISCORD.EXE", font=("Segoe UI", 8, "bold"), fg=CLR_TEXT_MUTED, bg=CLR_BG_CARD)
        lbl_path_title.place(x=15, y=10)

        detected_path = self.config_data.get("custom_discord_path", "")
        if not detected_path or not os.path.exists(detected_path):
            detected_path = find_discord_path()

        self.var_path = tk.StringVar(value=detected_path)

        entry_path = tk.Entry(
            card_path,
            textvariable=self.var_path,
            font=("Segoe UI", 9),
            bg=CLR_BG_INPUT,
            fg=CLR_TEXT_NORM,
            insertbackground=CLR_TEXT_WHITE,
            relief=tk.FLAT,
            highlightbackground=CLR_BORDER,
            highlightthickness=1
        )
        entry_path.place(x=15, y=30, width=320, height=30)

        btn_browse = tk.Button(
            card_path,
            text="Procurar...",
            font=("Segoe UI", 9),
            fg=CLR_TEXT_WHITE,
            bg=CLR_BTN_SEC,
            activebackground=CLR_BTN_SEC_HOV,
            activeforeground=CLR_TEXT_WHITE,
            relief=tk.FLAT,
            command=self.browse_file
        )
        btn_browse.place(x=348, y=29, width=90, height=32)

        btn_detect = tk.Button(
            card_path,
            text="Detectar",
            font=("Segoe UI", 9),
            fg=CLR_TEXT_WHITE,
            bg=CLR_BTN_SEC,
            activebackground=CLR_BTN_SEC_HOV,
            activeforeground=CLR_TEXT_WHITE,
            relief=tk.FLAT,
            command=self.detect_discord
        )
        btn_detect.place(x=445, y=29, width=80, height=32)

        lbl_path_tip = tk.Label(card_path, text="Detectado automaticamente no Registro do Windows / AppData.", font=("Segoe UI", 8), fg=CLR_TEXT_MUTED, bg=CLR_BG_CARD)
        lbl_path_tip.place(x=15, y=74)

        # 5. Main Action Button
        btn_launch = tk.Button(
            self,
            text="🚀  Abrir Discord",
            font=("Segoe UI", 12, "bold"),
            bg=CLR_BLURPLE,
            fg=CLR_TEXT_WHITE,
            activebackground=CLR_BLURPLE_HOV,
            activeforeground=CLR_TEXT_WHITE,
            relief=tk.FLAT,
            cursor="hand2",
            command=self.launch_discord
        )
        btn_launch.place(x=20, y=444, width=540, height=48)

    def get_args(self):
        host = self.var_host.get().strip()
        port = self.var_port.get().strip()
        args = []
        if host and port:
            clean_host = host
            protocol = "socks5://"
            if "://" in host:
                parts = host.split("://", 1)
                protocol = parts[0] + "://"
                clean_host = parts[1]
            args.append(f'--proxy-server="{protocol}{clean_host}:{port}"')
        if self.var_webrtc.get():
            args.append('--force-webrtc-ip-handling-policy=disable_non_proxied_udp')
        return args

    def browse_file(self):
        chosen = filedialog.askopenfilename(
            title="Selecione o executável do Discord",
            filetypes=[("Executáveis (*.exe)", "Discord.exe Update.exe *.exe"), ("Todos os Arquivos", "*.*")]
        )
        if chosen:
            self.var_path.set(chosen)

    def detect_discord(self):
        found = find_discord_path()
        if found:
            self.var_path.set(found)
            messagebox.showinfo("Discord Localizado", f"Discord encontrado em:\n{found}")
        else:
            messagebox.showwarning("Não encontrado", "Não foi possível localizar o Discord automaticamente. Por favor, use 'Procurar...' para selecionar.")

    def save_current_config(self):
        cfg = {
            "proxy_host": self.var_host.get().strip(),
            "proxy_port": self.var_port.get().strip(),
            "force_webrtc": self.var_webrtc.get(),
            "custom_discord_path": self.var_path.get().strip()
        }
        save_config(cfg)

    def on_close(self):
        self.save_current_config()
        self.destroy()

    def launch_discord(self):
        host = self.var_host.get().strip()
        port = self.var_port.get().strip()
        path = self.var_path.get().strip()

        if not host or not port:
            messagebox.showwarning("Campos Obrigatórios", "Por favor, preencha o Endereço de Proxy e a Porta.")
            return

        if not path or not os.path.exists(path):
            detected = find_discord_path()
            if detected and os.path.exists(detected):
                path = detected
                self.var_path.set(path)
            else:
                messagebox.showerror("Discord não encontrado", "Não foi possível encontrar o Discord. Use o botão 'Procurar...' para selecionar o arquivo Discord.exe.")
                return

        # 1. Check if Discord is running
        if is_discord_running():
            msg = "O Discord já está em execução.\n\nPara aplicar as novas configurações de Proxy e WebRTC, o Discord precisa ser fechado e reiniciado.\n\nDeseja fechar o Discord agora e abri-lo com o proxy configurado?"
            ans = messagebox.askyesno("Discord em Execução", msg)
            if ans:
                kill_discord()
            else:
                return

        # 2. Save settings next to exe (<exe_name>.config)
        self.save_current_config()
        args = self.get_args()

        # 3. Launch Discord completely detached
        try:
            flags = 0
            if sys.platform == 'win32':
                flags = subprocess.DETACHED_PROCESS | subprocess.CREATE_NEW_PROCESS_GROUP
            
            if os.path.basename(path).lower() == 'update.exe':
                cmd = [path, "--processStart", "Discord.exe"]
                if args:
                    cmd += ["--process-args", " ".join(args)]
                subprocess.Popen(cmd, cwd=os.path.dirname(path), creationflags=flags, close_fds=True)
            else:
                cmd = [path] + args
                subprocess.Popen(cmd, cwd=os.path.dirname(path), creationflags=flags, close_fds=True)
        except Exception as e:
            messagebox.showerror("Erro ao Iniciar", f"Falha ao executar o Discord:\n{e}")
            return

        # 4. Auto-close launcher immediately
        self.destroy()
        sys.exit(0)

if __name__ == '__main__':
    app = DiscordProxyLauncherApp()
    app.mainloop()
