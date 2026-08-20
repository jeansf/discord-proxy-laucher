# Discord Proxy Launcher - Windows 11

Aplicativo com tema oficial do Discord para Windows 11 que permite iniciar o Discord automaticamente configurado com servidor Proxy (SOCKS5/HTTP) e proteção contra vazamento de IP via WebRTC.

## Recursos

1. **Campos de Configuração**:
   * **Endereço do Proxy (IP/Host)**
   * **Porta**
   * **Checkbox**: **Forçar tunelamento WebRTC** (`--force-webrtc-ip-handling-policy=disable_non_proxied_udp`)
2. **Salvamento Automático**:
   * As configurações são salvas automaticamente sempre que você clica em **"🚀 Abrir Discord"** e ao fechar a janela.
3. **Localização Inteligente do Discord**:
   * Busca automática no Registro do Windows e em `%LOCALAPPDATA%\Discord\app-*\Discord.exe`.
   * Botão **"Procurar..."** para seleção manual de caminho caso necessário.
4. **Design e Ícone Customizado**:
   * Ícone exclusivo do Discord com o badge de escudo de segurança/proxy verde neon.
   * Totalmente integrado ao executável e barra de título com Dark Mode nativo do Windows 11.
5. **Executável Standalone**:
   * `DiscordProxyLauncher.exe` nativo de 64 bits (sem console, sem necessidade de instalar dependências no Windows).
