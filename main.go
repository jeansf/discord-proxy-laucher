package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// Win32 Constants
const (
	WS_OVERLAPPED    = 0x00000000
	WS_CAPTION       = 0x00C00000
	WS_SYSMENU       = 0x00080000
	WS_MINIMIZEBOX   = 0x00020000
	WS_VISIBLE       = 0x10000000
	WS_CHILD         = 0x40000000
	WS_TABSTOP       = 0x00010000
	WS_CLIPCHILDREN  = 0x02000000
	WS_CLIPSIBLINGS  = 0x04000000

	BS_PUSHBUTTON    = 0x00000000
	BS_DEFPUSHBUTTON = 0x00000001
	BS_CHECKBOX      = 0x00000002
	BS_AUTOCHECKBOX  = 0x00000003
	BS_OWNERDRAW     = 0x0000000B

	SS_NOTIFY        = 0x0100

	ES_LEFT          = 0x0000
	ES_AUTOHSCROLL   = 0x0080

	WM_DESTROY       = 0x0002
	WM_PAINT         = 0x000F
	WM_ERASEBKGND    = 0x0014
	WM_COMMAND       = 0x0111
	WM_SETFONT       = 0x0030
	WM_GETTEXT       = 0x000D
	WM_GETTEXTLENGTH = 0x000E
	WM_SETTEXT       = 0x000C
	WM_SETICON       = 0x0080
	WM_DRAWITEM      = 0x002B
	WM_CTLCOLORSTATIC = 0x0138
	WM_CTLCOLOREDIT   = 0x0133
	WM_CTLCOLORBTN    = 0x0135

	ICON_SMALL       = 0
	ICON_BIG         = 1
	IMAGE_ICON       = 1
	LR_DEFAULTCOLOR  = 0x0000
	LR_SHARED        = 0x8000

	BM_GETCHECK      = 0x00F0
	BM_SETCHECK      = 0x00F1
	BST_UNCHECKED    = 0
	BST_CHECKED      = 1

	SW_SHOW          = 5
	SW_HIDE          = 0

	MB_OK            = 0x00000000
	MB_YESNO         = 0x00000004
	MB_ICONINFORMATION = 0x00000040
	MB_ICONWARNING   = 0x00000030
	MB_ICONERROR     = 0x00000010
	MB_ICONQUESTION  = 0x00000020

	IDYES            = 6
	IDNO             = 7

	TH32CS_SNAPPROCESS = 0x00000002
	PROCESS_TERMINATE  = 0x0001

	HKEY_CURRENT_USER  = 0x80000001
	HKEY_LOCAL_MACHINE = 0x80000002
	KEY_READ           = 0x20019

	DWMWA_USE_IMMERSIVE_DARK_MODE     = 20
	DWMWA_USE_IMMERSIVE_DARK_MODE_OLD = 19

	ODS_SELECTED = 0x0001
	ODS_DISABLED = 0x0004
	ODS_FOCUS    = 0x0010

	DT_CENTER     = 0x0001
	DT_VCENTER    = 0x0004
	DT_SINGLELINE = 0x0020
	DT_LEFT       = 0x0000

	CREATE_NEW_PROCESS_GROUP = 0x00000200
	DETACHED_PROCESS         = 0x00000008
)

// Discord Palette Colors (BGR format for Win32 COLORREF)
const (
	COLOR_BG_MAIN     = 0x00383331 // #313338 (Discord dark main background)
	COLOR_BG_CARD     = 0x00312D2B // #2B2D31 (Discord card container background)
	COLOR_BG_INPUT    = 0x00221F1E // #1E1F22 (Discord text box background)
	COLOR_BORDER      = 0x0047413F // #3F4147 (Card / control border)
	COLOR_BLURPLE     = 0x00F26558 // #5865F2 (Discord Blurple primary button)
	COLOR_BLURPLE_HOV = 0x00C45247 // #4752C4 (Blurple hover)
	COLOR_BTN_SEC     = 0x0058504E // #4E5058 (Secondary button)
	COLOR_BTN_SEC_HOV = 0x006B6562 // #62656B (Secondary button hover)
	COLOR_TEXT_WHITE  = 0x00FFFFFF // #FFFFFF (White)
	COLOR_TEXT_NORM   = 0x00E1DEDB // #DBDEE1 (Primary light text)
	COLOR_TEXT_MUTED  = 0x00A49B94 // #949BA4 (Muted label text)
)

var (
	modUser32   = syscall.NewLazyDLL("user32.dll")
	modKernel32 = syscall.NewLazyDLL("kernel32.dll")
	modGdi32    = syscall.NewLazyDLL("gdi32.dll")
	modAdvapi32 = syscall.NewLazyDLL("advapi32.dll")
	modComdlg32 = syscall.NewLazyDLL("comdlg32.dll")
	modDwmapi   = syscall.NewLazyDLL("dwmapi.dll")

	procRegisterClassExW     = modUser32.NewProc("RegisterClassExW")
	procCreateWindowExW      = modUser32.NewProc("CreateWindowExW")
	procDefWindowProcW       = modUser32.NewProc("DefWindowProcW")
	procShowWindow           = modUser32.NewProc("ShowWindow")
	procUpdateWindow         = modUser32.NewProc("UpdateWindow")
	procGetMessageW          = modUser32.NewProc("GetMessageW")
	procTranslateMessage     = modUser32.NewProc("TranslateMessage")
	procDispatchMessageW     = modUser32.NewProc("DispatchMessageW")
	procPostQuitMessage      = modUser32.NewProc("PostQuitMessage")
	procSendMessageW         = modUser32.NewProc("SendMessageW")
	procSetWindowTextW       = modUser32.NewProc("SetWindowTextW")
	procGetWindowTextW       = modUser32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW = modUser32.NewProc("GetWindowTextLengthW")
	procMessageBoxW          = modUser32.NewProc("MessageBoxW")
	procGetSystemMetrics     = modUser32.NewProc("GetSystemMetrics")
	procLoadCursorW          = modUser32.NewProc("LoadCursorW")
	procLoadIconW            = modUser32.NewProc("LoadIconW")
	procLoadImageW           = modUser32.NewProc("LoadImageW")
	procDrawTextW            = modUser32.NewProc("DrawTextW")
	procDrawIconEx           = modUser32.NewProc("DrawIconEx")
	procBeginPaint           = modUser32.NewProc("BeginPaint")
	procEndPaint             = modUser32.NewProc("EndPaint")
	procFillRect             = modUser32.NewProc("FillRect")
	procGetClientRect        = modUser32.NewProc("GetClientRect")
	procSetProcessDPIAware   = modUser32.NewProc("SetProcessDPIAware")
	procAdjustWindowRectEx   = modUser32.NewProc("AdjustWindowRectEx")

	procGetModuleHandleW         = modKernel32.NewProc("GetModuleHandleW")
	procCreateProcessW           = modKernel32.NewProc("CreateProcessW")
	procCloseHandle              = modKernel32.NewProc("CloseHandle")
	procCreateToolhelp32Snapshot = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW          = modKernel32.NewProc("Process32FirstW")
	procProcess32NextW           = modKernel32.NewProc("Process32NextW")
	procOpenProcess              = modKernel32.NewProc("OpenProcess")
	procTerminateProcess         = modKernel32.NewProc("TerminateProcess")

	procCreateFontW          = modGdi32.NewProc("CreateFontW")
	procCreateSolidBrush     = modGdi32.NewProc("CreateSolidBrush")
	procCreatePen            = modGdi32.NewProc("CreatePen")
	procSetBkMode            = modGdi32.NewProc("SetBkMode")
	procSetBkColor           = modGdi32.NewProc("SetBkColor")
	procSetTextColor         = modGdi32.NewProc("SetTextColor")
	procSelectObject         = modGdi32.NewProc("SelectObject")
	procDeleteObject         = modGdi32.NewProc("DeleteObject")
	procRoundRect            = modGdi32.NewProc("RoundRect")

	procRegOpenKeyExW        = modAdvapi32.NewProc("RegOpenKeyExW")
	procRegQueryValueExW     = modAdvapi32.NewProc("RegQueryValueExW")
	procRegCloseKey          = modAdvapi32.NewProc("RegCloseKey")

	procGetOpenFileNameW     = modComdlg32.NewProc("GetOpenFileNameW")
	procDwmSetWindowAttribute = modDwmapi.NewProc("DwmSetWindowAttribute")
)

type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     syscall.Handle
	HIcon         syscall.Handle
	HCursor       syscall.Handle
	HbrBackground syscall.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       syscall.Handle
}

type MSG struct {
	HWnd    syscall.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type PAINTSTRUCT struct {
	Hdc         syscall.Handle
	FErase      int32
	RcPaint     RECT
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}

type DRAWITEMSTRUCT struct {
	CtlType    uint32
	CtlID      uint32
	ItemID     uint32
	ItemAction uint32
	ItemState  uint32
	HwndItem   syscall.Handle
	HDC        syscall.Handle
	RcItem     RECT
	ItemData   uintptr
}

type STARTUPINFOW struct {
	Cb              uint32
	LpReserved      *uint16
	LpDesktop       *uint16
	LpTitle         *uint16
	DwX             uint32
	DwY             uint32
	DwXSize         uint32
	DwYSize         uint32
	DwXCountChars   uint32
	DwYCountChars   uint32
	DwFillAttribute uint32
	DwFlags         uint32
	WShowWindow     uint16
	CbReserved2     uint16
	LpReserved2     *byte
	HStdInput       syscall.Handle
	HStdOutput      syscall.Handle
	HStdError       syscall.Handle
}

type PROCESS_INFORMATION struct {
	HProcess    syscall.Handle
	HThread     syscall.Handle
	DwProcessId uint32
	DwThreadId  uint32
}

type PROCESSENTRY32W struct {
	DwSize              uint32
	CntUsage            uint32
	Th32ProcessID       uint32
	Th32DefaultHeapID   uintptr
	Th32ModuleID        uint32
	CntThreads          uint32
	Th32ParentProcessID uint32
	PcPriClassBase      int32
	DwFlags             uint32
	SzExeFile           [260]uint16
}

type OPENFILENAMEW struct {
	LStructSize       uint32
	HwndOwner         syscall.Handle
	HInstance         syscall.Handle
	LpstrFilter       *uint16
	LpstrCustomFilter *uint16
	NMaxCustFilter    uint32
	NFilterIndex      uint32
	LpstrFile         *uint16
	NMaxFile          uint32
	LpstrFileTitle    *uint16
	NMaxFileTitle     uint32
	LpstrInitialDir   *uint16
	LpstrTitle        *uint16
	Flags             uint32
	NFileOffset       uint16
	NFileExtension    uint16
	LpstrDefExt       *uint16
	LCustData         uintptr
	LpfnHook          uintptr
	LpTemplateName    *uint16
}

type Config struct {
	ProxyHost         string `json:"proxy_host"`
	ProxyPort         string `json:"proxy_port"`
	ForceWebRTC       bool   `json:"force_webrtc"`
	CustomDiscordPath string `json:"custom_discord_path"`
}

const (
	ID_PROXY_HOST     = 101
	ID_PROXY_PORT     = 102
	ID_WEBRTC_CHK     = 103
	ID_WEBRTC_LBL     = 104
	ID_DISCORD_PATH   = 105
	ID_BTN_BROWSE     = 106
	ID_BTN_DETECT     = 107
	ID_BTN_LAUNCH     = 108
)

var (
	hMainWnd       syscall.Handle
	hHeaderTitle   syscall.Handle
	hHeaderSub     syscall.Handle
	hEditHost      syscall.Handle
	hEditPort      syscall.Handle
	hChkWebRTC     syscall.Handle
	hLblWebRTC     syscall.Handle
	hEditPath      syscall.Handle
	hAppIcon       syscall.Handle

	hFontHeader    uintptr
	hFontSub       uintptr
	hFontLabel     uintptr
	hFontNormal    uintptr
	hFontBold      uintptr
	hFontBtn       uintptr

	hBrushMainBg   uintptr
	hBrushCardBg   uintptr
	hBrushInputBg  uintptr
	hBrushBlurple  uintptr
	hBrushBlurpleD uintptr
	hBrushSecBtn   uintptr
	hBrushSecBtnD  uintptr
	hPenBorder     uintptr
	hPenNull       uintptr

	currentConfig  Config
)

func utf16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func getControlText(hCtrl syscall.Handle) string {
	lenRes, _, _ := procGetWindowTextLengthW.Call(uintptr(hCtrl))
	length := int(lenRes)
	if length == 0 {
		return ""
	}
	buf := make([]uint16, length+1)
	procGetWindowTextW.Call(uintptr(hCtrl), uintptr(unsafe.Pointer(&buf[0])), uintptr(length+1))
	return syscall.UTF16ToString(buf)
}

func setControlText(hCtrl syscall.Handle, text string) {
	procSetWindowTextW.Call(uintptr(hCtrl), uintptr(unsafe.Pointer(utf16Ptr(text))))
}

func isChecked(hCtrl syscall.Handle) bool {
	res, _, _ := procSendMessageW.Call(uintptr(hCtrl), BM_GETCHECK, 0, 0)
	return res == BST_CHECKED
}

func setChecked(hCtrl syscall.Handle, checked bool) {
	val := uintptr(BST_UNCHECKED)
	if checked {
		val = uintptr(BST_CHECKED)
	}
	procSendMessageW.Call(uintptr(hCtrl), BM_SETCHECK, val, 0)
}

func showMessage(hwnd syscall.Handle, text, title string, flags uint32) int {
	res, _, _ := procMessageBoxW.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(utf16Ptr(text))),
		uintptr(unsafe.Pointer(utf16Ptr(title))),
		uintptr(flags),
	)
	return int(res)
}

func getConfigPath() string {
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		name := strings.TrimSuffix(filepath.Base(exe), filepath.Ext(exe))
		return filepath.Join(dir, name+".config")
	}
	return "DiscordProxyLauncher.config"
}

func loadConfig() Config {
	cfg := Config{
		ProxyHost:   "",
		ProxyPort:   "",
		ForceWebRTC: true,
	}

	path := getConfigPath()
	data, err := os.ReadFile(path)
	if err == nil {
		var loaded Config
		if err := json.Unmarshal(data, &loaded); err == nil {
			cfg.ProxyHost = loaded.ProxyHost
			cfg.ProxyPort = loaded.ProxyPort
			cfg.ForceWebRTC = loaded.ForceWebRTC
			cfg.CustomDiscordPath = loaded.CustomDiscordPath
			return cfg
		}
	}

	exe, err := os.Executable()
	if err == nil {
		oldPath := filepath.Join(filepath.Dir(exe), "config.json")
		if oldData, err2 := os.ReadFile(oldPath); err2 == nil {
			var loaded Config
			if err3 := json.Unmarshal(oldData, &loaded); err3 == nil {
				cfg.ProxyHost = loaded.ProxyHost
				cfg.ProxyPort = loaded.ProxyPort
				cfg.ForceWebRTC = loaded.ForceWebRTC
				cfg.CustomDiscordPath = loaded.CustomDiscordPath
			}
		}
	}
	return cfg
}

func saveConfig(cfg Config) {
	defer func() { _ = recover() }()
	path := getConfigPath()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err == nil {
		_ = os.WriteFile(path, data, 0666)
	}
}

func saveCurrentState() {
	if hEditHost == 0 || hEditPort == 0 {
		return
	}
	host := strings.TrimSpace(getControlText(hEditHost))
	port := strings.TrimSpace(getControlText(hEditPort))
	webrtc := isChecked(hChkWebRTC)
	path := strings.TrimSpace(getControlText(hEditPath))
	cfg := Config{
		ProxyHost:         host,
		ProxyPort:         port,
		ForceWebRTC:       webrtc,
		CustomDiscordPath: path,
	}
	saveConfig(cfg)
}

func getRunningDiscordPIDs() []uint32 {
	hSnap, _, _ := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPPROCESS, 0)
	if hSnap == uintptr(syscall.InvalidHandle) || hSnap == 0 {
		return nil
	}
	defer procCloseHandle.Call(hSnap)

	var pe PROCESSENTRY32W
	pe.DwSize = uint32(unsafe.Sizeof(pe))

	var pids []uint32
	res, _, _ := procProcess32FirstW.Call(hSnap, uintptr(unsafe.Pointer(&pe)))
	for res != 0 {
		exeName := syscall.UTF16ToString(pe.SzExeFile[:])
		if strings.EqualFold(exeName, "discord.exe") ||
			strings.EqualFold(exeName, "discordcanary.exe") ||
			strings.EqualFold(exeName, "discordptb.exe") ||
			strings.EqualFold(exeName, "discorddevelopment.exe") {
			pids = append(pids, pe.Th32ProcessID)
		}
		res, _, _ = procProcess32NextW.Call(hSnap, uintptr(unsafe.Pointer(&pe)))
	}
	return pids
}

func killDiscordProcesses(pids []uint32) {
	for _, pid := range pids {
		hProc, _, _ := procOpenProcess.Call(PROCESS_TERMINATE, 0, uintptr(pid))
		if hProc != 0 && hProc != uintptr(syscall.InvalidHandle) {
			procTerminateProcess.Call(hProc, 0)
			procCloseHandle.Call(hProc)
		}
	}
	time.Sleep(350 * time.Millisecond)
}

func readRegString(hKeyRoot uintptr, subKey, valueName string) (string, error) {
	var hKey uintptr
	res, _, _ := procRegOpenKeyExW.Call(
		hKeyRoot,
		uintptr(unsafe.Pointer(utf16Ptr(subKey))),
		0,
		KEY_READ,
		uintptr(unsafe.Pointer(&hKey)),
	)
	if res != 0 {
		return "", fmt.Errorf("RegOpenKeyExW failed: %d", res)
	}
	defer procRegCloseKey.Call(hKey)

	var valType uint32
	var bufSize uint32 = 1024
	buf := make([]uint16, 512)

	res, _, _ = procRegQueryValueExW.Call(
		hKey,
		uintptr(unsafe.Pointer(utf16Ptr(valueName))),
		0,
		uintptr(unsafe.Pointer(&valType)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&bufSize)),
	)
	if res != 0 {
		return "", fmt.Errorf("RegQueryValueExW failed: %d", res)
	}
	return syscall.UTF16ToString(buf), nil
}

func findDiscordPath() string {
	cmdStr, err := readRegString(HKEY_CURRENT_USER, `Software\Classes\discord\shell\open\command`, "")
	if err == nil && cmdStr != "" {
		path := extractPathFromCommand(cmdStr)
		if path != "" {
			if strings.EqualFold(filepath.Base(path), "discord.exe") && fileExists(path) {
				return path
			}
			dir := filepath.Dir(path)
			discordExe := findDiscordInFolder(dir)
			if discordExe != "" {
				return discordExe
			}
			if fileExists(path) {
				return path
			}
		}
	}

	displayIcon, err := readRegString(HKEY_CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Uninstall\Discord`, "DisplayIcon")
	if err == nil && displayIcon != "" {
		clean := cleanFilePath(displayIcon)
		if strings.EqualFold(filepath.Base(clean), "discord.exe") && fileExists(clean) {
			return clean
		}
		dir := filepath.Dir(clean)
		discordExe := findDiscordInFolder(dir)
		if discordExe != "" {
			return discordExe
		}
	}

	installLoc, err := readRegString(HKEY_CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Uninstall\Discord`, "InstallLocation")
	if err == nil && installLoc != "" {
		discordExe := findDiscordInFolder(installLoc)
		if discordExe != "" {
			return discordExe
		}
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		candidates := []string{
			filepath.Join(localAppData, "Discord"),
			filepath.Join(localAppData, "DiscordCanary"),
			filepath.Join(localAppData, "DiscordPTB"),
			filepath.Join(localAppData, "DiscordDevelopment"),
		}
		for _, dir := range candidates {
			discordExe := findDiscordInFolder(dir)
			if discordExe != "" {
				return discordExe
			}
			updateExe := filepath.Join(dir, "Update.exe")
			if fileExists(updateExe) {
				return updateExe
			}
		}
	}

	progFiles := []string{
		os.Getenv("ProgramFiles"),
		os.Getenv("ProgramFiles(x86)"),
		`C:\Program Files`,
		`C:\Program Files (x86)`,
	}
	for _, pf := range progFiles {
		if pf != "" {
			p := filepath.Join(pf, "Discord", "Discord.exe")
			if fileExists(p) {
				return p
			}
		}
	}

	return ""
}

func findDiscordInFolder(dir string) string {
	if !dirExists(dir) {
		return ""
	}
	direct := filepath.Join(dir, "Discord.exe")
	if fileExists(direct) {
		return direct
	}

	entries, err := os.ReadDir(dir)
	if err == nil {
		var appDirs []string
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "app-") {
				appDirs = append(appDirs, entry.Name())
			}
		}
		sort.Slice(appDirs, func(i, j int) bool {
			return appDirs[i] > appDirs[j]
		})
		for _, d := range appDirs {
			target := filepath.Join(dir, d, "Discord.exe")
			if fileExists(target) {
				return target
			}
		}
	}
	return ""
}

func cleanFilePath(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.Trim(s, "\"")
	if idx := strings.Index(s, ","); idx > 0 && !strings.Contains(s[idx:], `\`) {
		s = s[:idx]
	}
	return strings.TrimSpace(s)
}

func extractPathFromCommand(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	if strings.HasPrefix(cmd, "\"") {
		parts := strings.Split(cmd[1:], "\"")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	fields := strings.Fields(cmd)
	if len(fields) > 0 {
		return fields[0]
	}
	return cmd
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

func dirExists(path string) bool {
	if path == "" {
		return false
	}
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

func buildLaunchArgs(proxyHost, proxyPort string, forceWebRTC bool) []string {
	var args []string
	proxyHost = strings.TrimSpace(proxyHost)
	proxyPort = strings.TrimSpace(proxyPort)
	if proxyHost != "" && proxyPort != "" {
		hostClean := proxyHost
		protocol := "socks5://"
		if strings.Contains(proxyHost, "://") {
			parts := strings.SplitN(proxyHost, "://", 2)
			protocol = parts[0] + "://"
			hostClean = parts[1]
		}
		args = append(args, fmt.Sprintf("--proxy-server=\"%s%s:%s\"", protocol, hostClean, proxyPort))
	}
	if forceWebRTC {
		args = append(args, "--force-webrtc-ip-handling-policy=disable_non_proxied_udp")
	}
	return args
}

func browseForDiscord(hwnd syscall.Handle) string {
	var ofn OPENFILENAMEW
	ofn.LStructSize = uint32(unsafe.Sizeof(ofn))
	ofn.HwndOwner = hwnd
	
	filter := "Executáveis do Discord (*.exe)\x00Discord.exe;Update.exe\x00Todos os Arquivos (*.*)\x00*.*\x00\x00"
	filterBuf := make([]uint16, len(filter))
	for i, c := range filter {
		filterBuf[i] = uint16(c)
	}
	ofn.LpstrFilter = &filterBuf[0]

	fileBuf := make([]uint16, 1024)
	ofn.LpstrFile = &fileBuf[0]
	ofn.NMaxFile = 1024
	ofn.LpstrTitle = utf16Ptr("Selecione o executável do Discord")
	ofn.Flags = 0x00000800 | 0x00000008

	ret, _, _ := procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if ret != 0 {
		return syscall.UTF16ToString(fileBuf)
	}
	return ""
}

func launchDiscord(hwnd syscall.Handle) {
	host := strings.TrimSpace(getControlText(hEditHost))
	port := strings.TrimSpace(getControlText(hEditPort))
	webrtc := isChecked(hChkWebRTC)
	path := strings.TrimSpace(getControlText(hEditPath))

	if host == "" || port == "" {
		showMessage(hwnd, "Por favor, preencha o Endereço de Proxy e a Porta.", "Campos Obrigatórios", MB_OK|MB_ICONWARNING)
		return
	}

	if path == "" || !fileExists(path) {
		detected := findDiscordPath()
		if detected != "" && fileExists(detected) {
			path = detected
			setControlText(hEditPath, path)
		} else {
			showMessage(hwnd, "Não foi possível localizar o Discord automaticamente.\nPor favor, utilize o botão 'Procurar...' para selecionar o executável do Discord.", "Discord não encontrado", MB_OK|MB_ICONERROR)
			return
		}
	}

	// 1. Check if Discord is currently running
	runningPIDs := getRunningDiscordPIDs()
	if len(runningPIDs) > 0 {
		msg := "O Discord já está em execução.\n\nPara aplicar as novas configurações de Proxy e WebRTC, o Discord precisa ser fechado e reiniciado.\n\nDeseja fechar o Discord agora e abri-lo com o proxy configurado?"
		ans := showMessage(hwnd, msg, "Discord em Execução", MB_YESNO|MB_ICONQUESTION)
		if ans == IDYES {
			killDiscordProcesses(runningPIDs)
		} else {
			return
		}
	}

	// 2. Save config to .config in executable directory
	cfg := Config{
		ProxyHost:         host,
		ProxyPort:         port,
		ForceWebRTC:       webrtc,
		CustomDiscordPath: path,
	}
	saveConfig(cfg)

	args := buildLaunchArgs(host, port, webrtc)

	// 3. Prepare writable command-line buffer for CreateProcessW
	var fullCmdLine string
	if strings.EqualFold(filepath.Base(path), "update.exe") {
		fullCmdLine = fmt.Sprintf("\"%s\" --processStart Discord.exe", path)
		if len(args) > 0 {
			fullCmdLine += fmt.Sprintf(" --process-args \"%s\"", strings.Join(args, " "))
		}
	} else {
		fullCmdLine = fmt.Sprintf("\"%s\" %s", path, strings.Join(args, " "))
	}

	var si STARTUPINFOW
	si.Cb = uint32(unsafe.Sizeof(si))
	var pi PROCESS_INFORMATION

	workingDir := filepath.Dir(path)
	cmdBuf := syscall.StringToUTF16(fullCmdLine)
	dirBuf := syscall.StringToUTF16(workingDir)

	res, _, _ := procCreateProcessW.Call(
		0,
		uintptr(unsafe.Pointer(&cmdBuf[0])), // mutable memory buffer
		0,
		0,
		0, // bInheritHandles = FALSE
		CREATE_NEW_PROCESS_GROUP|DETACHED_PROCESS,
		0,
		uintptr(unsafe.Pointer(&dirBuf[0])),
		uintptr(unsafe.Pointer(&si)),
		uintptr(unsafe.Pointer(&pi)),
	)

	if res != 0 {
		procCloseHandle.Call(uintptr(pi.HProcess))
		procCloseHandle.Call(uintptr(pi.HThread))
	}

	// 4. Immediately exit cleanly
	os.Exit(0)
}

func createFont(name string, size int, bold bool) uintptr {
	weight := 400
	if bold {
		weight = 700
	}
	hFont, _, _ := procCreateFontW.Call(
		uintptr(-size), 0, 0, 0,
		uintptr(weight),
		0, 0, 0, 1, 0, 0, 0, 0,
		uintptr(unsafe.Pointer(utf16Ptr(name))),
	)
	return hFont
}

func drawCustomButton(dis *DRAWITEMSTRUCT, text string, isPrimary bool) {
	hdc := uintptr(dis.HDC)
	rect := dis.RcItem

	bgBrush := hBrushBlurple
	if isPrimary {
		if dis.ItemState&ODS_SELECTED != 0 {
			bgBrush = hBrushBlurpleD
		}
	} else {
		bgBrush = hBrushSecBtn
		if dis.ItemState&ODS_SELECTED != 0 {
			bgBrush = hBrushSecBtnD
		}
	}

	procSelectObject.Call(hdc, bgBrush)
	procSelectObject.Call(hdc, hPenNull)
	procRoundRect.Call(hdc, uintptr(rect.Left), uintptr(rect.Top), uintptr(rect.Right), uintptr(rect.Bottom), 8, 8)

	procSetBkMode.Call(hdc, 1)
	procSetTextColor.Call(hdc, uintptr(COLOR_TEXT_WHITE))
	procSelectObject.Call(hdc, hFontBtn)

	r := rect
	textPtr := utf16Ptr(text)
	// Pass -1 (^uintptr(0)) so DrawTextW reads full null-terminated UTF-16 text without emoji surrogate pair truncation!
	procDrawTextW.Call(
		hdc,
		uintptr(unsafe.Pointer(textPtr)),
		^uintptr(0),
		uintptr(unsafe.Pointer(&r)),
		DT_CENTER|DT_VCENTER|DT_SINGLELINE,
	)
}

// Global static callback: 0 heap allocations, 0 closures, 0 re-entrant lock calls
func wndProc(hwnd, msg, wParam, lParam uintptr) (result uintptr) {
	defer func() {
		if r := recover(); r != nil {
			result = 0
		}
	}()

	switch uint32(msg) {
	case WM_ERASEBKGND:
		return 1

	case WM_PAINT:
		var ps PAINTSTRUCT
		hdcRes, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		hdc := uintptr(hdcRes)
		if hdc != 0 {
			// Fill full client background
			var rc RECT
			procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
			procFillRect.Call(hdc, uintptr(unsafe.Pointer(&rc)), hBrushMainBg)

			// Select card background and border
			procSelectObject.Call(hdc, hBrushCardBg)
			procSelectObject.Call(hdc, hPenBorder)

			// Draw the 3 card containers (Left: 20px, Right: 560px -> Width: 540px)
			procRoundRect.Call(hdc, 20, 68, 560, 180, 12, 12)
			procRoundRect.Call(hdc, 20, 194, 560, 298, 12, 12)
			procRoundRect.Call(hdc, 20, 312, 560, 424, 12, 12)

			// Draw App Icon in header
			if hAppIcon != 0 {
				procDrawIconEx.Call(hdc, 20, 14, uintptr(hAppIcon), 40, 40, 0, 0, 3)
			}

			procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		}
		return 0

	case WM_DRAWITEM:
		dis := (*DRAWITEMSTRUCT)(unsafe.Pointer(lParam))
		switch dis.CtlID {
		case ID_BTN_LAUNCH:
			drawCustomButton(dis, "🚀  Abrir Discord", true)
			return 1
		case ID_BTN_BROWSE:
			drawCustomButton(dis, "Procurar...", false)
			return 1
		case ID_BTN_DETECT:
			drawCustomButton(dis, "Detectar", false)
			return 1
		}

	case WM_CTLCOLORSTATIC:
		hdc := wParam
		procSetBkMode.Call(hdc, 1)
		if lParam == uintptr(hHeaderTitle) || lParam == uintptr(hHeaderSub) {
			procSetBkColor.Call(hdc, uintptr(COLOR_BG_MAIN))
			procSetTextColor.Call(hdc, uintptr(COLOR_TEXT_NORM))
			return hBrushMainBg
		}
		procSetBkColor.Call(hdc, uintptr(COLOR_BG_CARD))
		procSetTextColor.Call(hdc, uintptr(COLOR_TEXT_NORM))
		return hBrushCardBg

	case WM_CTLCOLOREDIT:
		hdc := wParam
		procSetBkColor.Call(hdc, uintptr(COLOR_BG_INPUT))
		procSetTextColor.Call(hdc, uintptr(COLOR_TEXT_NORM))
		return hBrushInputBg

	case WM_CTLCOLORBTN:
		hdc := wParam
		procSetBkMode.Call(hdc, 1)
		procSetTextColor.Call(hdc, uintptr(COLOR_TEXT_NORM))
		return hBrushCardBg

	case WM_COMMAND:
		ctrlID := int(wParam & 0xFFFF)
		switch ctrlID {
		case ID_BTN_LAUNCH:
			launchDiscord(syscall.Handle(hwnd))
		case ID_WEBRTC_LBL:
			checked := isChecked(hChkWebRTC)
			setChecked(hChkWebRTC, !checked)
		case ID_BTN_BROWSE:
			selected := browseForDiscord(syscall.Handle(hwnd))
			if selected != "" {
				setControlText(hEditPath, selected)
			}
		case ID_BTN_DETECT:
			detected := findDiscordPath()
			if detected != "" {
				setControlText(hEditPath, detected)
				showMessage(syscall.Handle(hwnd), fmt.Sprintf("Discord localizado com sucesso em:\n%s", detected), "Discord Localizado", MB_OK|MB_ICONINFORMATION)
			} else {
				showMessage(syscall.Handle(hwnd), "Discord não foi localizado automaticamente no Registro nem no LocalAppData.", "Não encontrado", MB_OK|MB_ICONWARNING)
			}
		}
		return 0

	case WM_DESTROY:
		saveCurrentState()
		procPostQuitMessage.Call(0)
		return 0
	}

	res, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return res
}

func main() {
	runtime.LockOSThread()

	if procSetProcessDPIAware.Find() == nil {
		procSetProcessDPIAware.Call()
	}

	currentConfig = loadConfig()

	discordPath := currentConfig.CustomDiscordPath
	if discordPath == "" || !fileExists(discordPath) {
		detected := findDiscordPath()
		if detected != "" {
			discordPath = detected
		}
	}

	hInstRes, _, _ := procGetModuleHandleW.Call(0)
	hInstance := syscall.Handle(hInstRes)

	// Load MAINICON from embedded resource
	iconRes, _, _ := procLoadImageW.Call(
		uintptr(hInstance),
		1,
		IMAGE_ICON,
		0, 0,
		LR_DEFAULTCOLOR|LR_SHARED,
	)
	if iconRes != 0 {
		hAppIcon = syscall.Handle(iconRes)
	} else {
		stdIcon, _, _ := procLoadIconW.Call(0, 32512)
		hAppIcon = syscall.Handle(stdIcon)
	}

	className := "DiscordProxyLauncherWndClass"
	classNamePtr := utf16Ptr(className)

	hFontHeader = createFont("Segoe UI", 21, true)
	hFontSub = createFont("Segoe UI", 13, false)
	hFontLabel = createFont("Segoe UI", 12, true)
	hFontBold = createFont("Segoe UI", 14, true)
	hFontNormal = createFont("Segoe UI", 14, false)
	hFontBtn = createFont("Segoe UI", 15, true)

	brushRes, _, _ := procCreateSolidBrush.Call(uintptr(COLOR_BG_MAIN))
	hBrushMainBg = brushRes
	brushRes, _, _ = procCreateSolidBrush.Call(uintptr(COLOR_BG_CARD))
	hBrushCardBg = brushRes
	brushRes, _, _ = procCreateSolidBrush.Call(uintptr(COLOR_BG_INPUT))
	hBrushInputBg = brushRes
	brushRes, _, _ = procCreateSolidBrush.Call(uintptr(COLOR_BLURPLE))
	hBrushBlurple = brushRes
	brushRes, _, _ = procCreateSolidBrush.Call(uintptr(COLOR_BLURPLE_HOV))
	hBrushBlurpleD = brushRes
	brushRes, _, _ = procCreateSolidBrush.Call(uintptr(COLOR_BTN_SEC))
	hBrushSecBtn = brushRes
	brushRes, _, _ = procCreateSolidBrush.Call(uintptr(COLOR_BTN_SEC_HOV))
	hBrushSecBtnD = brushRes

	penRes, _, _ := procCreatePen.Call(0, 1, uintptr(COLOR_BORDER))
	hPenBorder = penRes
	penRes, _, _ = procCreatePen.Call(5, 0, 0)
	hPenNull = penRes

	cursorRes, _, _ := procLoadCursorW.Call(0, 32512)

	var wc WNDCLASSEXW
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	wc.Style = 0x0003
	wc.LpfnWndProc = syscall.NewCallback(wndProc)
	wc.HInstance = hInstance
	wc.HbrBackground = syscall.Handle(hBrushMainBg)
	wc.LpszClassName = classNamePtr
	wc.HCursor = syscall.Handle(cursorRes)
	wc.HIcon = hAppIcon
	wc.HIconSm = hAppIcon

	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	// Desired client area: EXACTLY 580px width x 515px height
	clientW := int32(580)
	clientH := int32(515)

	winStyle := uint32(WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU | WS_MINIMIZEBOX | WS_CLIPCHILDREN)

	var winRect RECT
	winRect.Left = 0
	winRect.Top = 0
	winRect.Right = clientW
	winRect.Bottom = clientH

	procAdjustWindowRectEx.Call(
		uintptr(unsafe.Pointer(&winRect)),
		uintptr(winStyle),
		0,
		0,
	)

	outerW := winRect.Right - winRect.Left
	outerH := winRect.Bottom - winRect.Top

	scrW, _, _ := procGetSystemMetrics.Call(0)
	scrH, _, _ := procGetSystemMetrics.Call(1)
	posX := (int32(scrW) - outerW) / 2
	posY := (int32(scrH) - outerH) / 2

	hMainRes, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(classNamePtr)),
		uintptr(unsafe.Pointer(utf16Ptr("Discord Proxy Launcher"))),
		uintptr(winStyle),
		uintptr(posX), uintptr(posY), uintptr(outerW), uintptr(outerH),
		0, 0, uintptr(hInstance), 0,
	)
	hMainWnd = syscall.Handle(hMainRes)

	if hAppIcon != 0 {
		procSendMessageW.Call(uintptr(hMainWnd), WM_SETICON, ICON_BIG, uintptr(hAppIcon))
		procSendMessageW.Call(uintptr(hMainWnd), WM_SETICON, ICON_SMALL, uintptr(hAppIcon))
	}

	darkModeVal := int32(1)
	procDwmSetWindowAttribute.Call(
		uintptr(hMainWnd),
		DWMWA_USE_IMMERSIVE_DARK_MODE,
		uintptr(unsafe.Pointer(&darkModeVal)),
		4,
	)
	procDwmSetWindowAttribute.Call(
		uintptr(hMainWnd),
		DWMWA_USE_IMMERSIVE_DARK_MODE_OLD,
		uintptr(unsafe.Pointer(&darkModeVal)),
		4,
	)

	createControl := func(className, text string, style uint32, x, y, w, h int32, id int) syscall.Handle {
		res, _, _ := procCreateWindowExW.Call(
			0,
			uintptr(unsafe.Pointer(utf16Ptr(className))),
			uintptr(unsafe.Pointer(utf16Ptr(text))),
			uintptr(WS_CHILD|WS_VISIBLE|WS_CLIPSIBLINGS|style),
			uintptr(x), uintptr(y), uintptr(w), uintptr(h),
			uintptr(hMainWnd), uintptr(id), uintptr(hInstance), 0,
		)
		hCtrl := syscall.Handle(res)
		procSendMessageW.Call(uintptr(hCtrl), WM_SETFONT, hFontNormal, 1)
		return hCtrl
	}

	// 1. Header (Left: 20px icon, 68px text. Right edge: 560px)
	hHeaderTitle = createControl("STATIC", "DISCORD PROXY LAUNCHER", 0, 68, 14, 492, 24, 0)
	procSendMessageW.Call(uintptr(hHeaderTitle), WM_SETFONT, hFontHeader, 1)

	hHeaderSub = createControl("STATIC", "Inicie o Discord com proxy SOCKS5 e proteção contra vazamento WebRTC", 0, 68, 38, 492, 20, 0)
	procSendMessageW.Call(uintptr(hHeaderSub), WM_SETFONT, hFontSub, 1)

	// 2. Card 1: Proxy Settings (Left 20px, Right 560px -> Width 540px)
	hLblProxyH := createControl("STATIC", "ENDEREÇO DO PROXY (IP OU HOST)", 0, 35, 78, 300, 18, 0)
	procSendMessageW.Call(uintptr(hLblProxyH), WM_SETFONT, hFontLabel, 1)
	hEditHost = createControl("EDIT", currentConfig.ProxyHost, WS_TABSTOP|ES_LEFT|ES_AUTOHSCROLL, 35, 98, 330, 28, ID_PROXY_HOST)

	hLblProxyP := createControl("STATIC", "PORTA", 0, 380, 78, 120, 18, 0)
	procSendMessageW.Call(uintptr(hLblProxyP), WM_SETFONT, hFontLabel, 1)
	hEditPort = createControl("EDIT", currentConfig.ProxyPort, WS_TABSTOP|ES_LEFT|ES_AUTOHSCROLL, 380, 98, 165, 28, ID_PROXY_PORT)

	hLblEx := createControl("STATIC", "Protocolo padrão: socks5:// (ex: 127.0.0.1 porta 1080)", 0, 35, 142, 510, 20, 0)
	procSendMessageW.Call(uintptr(hLblEx), WM_SETFONT, hFontSub, 1)

	// 3. Card 2: WebRTC (Left 20px, Right 560px -> Width 540px)
	hChkWebRTC = createControl("BUTTON", "", BS_AUTOCHECKBOX|WS_TABSTOP, 35, 206, 20, 20, ID_WEBRTC_CHK)
	setChecked(hChkWebRTC, currentConfig.ForceWebRTC)

	hLblWebRTC = createControl("STATIC", "Forçar tunelamento WebRTC", SS_NOTIFY, 62, 206, 483, 20, ID_WEBRTC_LBL)
	procSendMessageW.Call(uintptr(hLblWebRTC), WM_SETFONT, hFontBold, 1)

	hLblDesc := createControl("STATIC", "Bloqueia vazamento de IP real desativando conexões UDP diretas (--force-webrtc-ip-handling-policy=disable_non_proxied_udp).", 0, 62, 232, 483, 40, 0)
	procSendMessageW.Call(uintptr(hLblDesc), WM_SETFONT, hFontSub, 1)

	// 4. Card 3: Discord Location (Left 20px, Right 560px -> Width 540px)
	hLblPath := createControl("STATIC", "LOCALIZAÇÃO DO DISCORD.EXE", 0, 35, 324, 300, 18, 0)
	procSendMessageW.Call(uintptr(hLblPath), WM_SETFONT, hFontLabel, 1)

	hEditPath = createControl("EDIT", discordPath, WS_TABSTOP|ES_LEFT|ES_AUTOHSCROLL, 35, 344, 320, 28, ID_DISCORD_PATH)

	createControl("BUTTON", "Procurar...", BS_OWNERDRAW|WS_TABSTOP, 368, 343, 90, 30, ID_BTN_BROWSE)
	createControl("BUTTON", "Detectar", BS_OWNERDRAW|WS_TABSTOP, 465, 343, 80, 30, ID_BTN_DETECT)

	hLblPathTip := createControl("STATIC", "Detectado automaticamente no Registro do Windows / AppData.", 0, 35, 386, 510, 20, 0)
	procSendMessageW.Call(uintptr(hLblPathTip), WM_SETFONT, hFontSub, 1)

	// 5. Main Action Button: Left 20px, Width 540px (Right edge 560px -> 20px right margin!)
	createControl("BUTTON", "🚀  Abrir Discord", BS_OWNERDRAW|WS_TABSTOP, 20, 444, 540, 48, ID_BTN_LAUNCH)

	procShowWindow.Call(uintptr(hMainWnd), SW_SHOW)
	procUpdateWindow.Call(uintptr(hMainWnd))

	var msg MSG
	for {
		res, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(res) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}
