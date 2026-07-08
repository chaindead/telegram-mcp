[![](https://badge.mcpx.dev?type=server 'MCP Server')](https://github.com/punkpeye/awesome-mcp-servers?tab=readme-ov-file#communication)
[![](https://img.shields.io/badge/OS_Agnostic-Works_Everywhere-purple)](https://github.com/chaindead/telegram-mcp?tab=readme-ov-file#installation)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Visitors](https://api.visitorbadge.io/api/visitors?path=https%3A%2F%2Fgithub.com%2Fchaindead%2Ftelegram-mcp&label=Visitors&labelColor=%23d9e3f0&countColor=%23697689&style=flat&labelStyle=none)](https://visitorbadge.io/status?path=https%3A%2F%2Fgithub.com%2Fchaindead%2Ftelegram-mcp)

# Telegram MCP server

The server is a bridge between the Telegram API and the AI assistants and is based on the [Model Context Protocol](https://modelcontextprotocol.io).

> [!IMPORTANT]
> Ensure that you have read and understood the [Telegram API Terms of Service](https://core.telegram.org/api/terms) before using this server.
> Any misuse of the Telegram API may result in the suspension of your account.

## Table of Contents
- [What is MCP?](#what-is-mcp)
- [What does this server do?](#what-does-this-server-do)
  - [Capabilities](#capabilities)
  - [Prompt examples](#prompt-examples)
    - [Message Management](#message-management)
    - [Organization](#organization)
    - [Communication](#communication)
- [Installation](#installation)
  - [Homebrew](#homebrew)
  - [NPX](#npx)
  - [From Releases](#from-releases)
    - [MacOS](#macos)
    - [Linux](#linux)
    - [Windows](#windows)
  - [From Source](#from-source)
- [Configuration](#configuration)
  - [Authorization](#authorization)
  - [Client Configuration](#client-configuration)
  - [JSON Schema Version](#json-schema-version)
- [Star History](#star-history)

## What is MCP?

The Model Context Protocol (MCP) is a system that lets AI apps, like Claude Desktop or Cursor, connect to external tools and data sources. It gives a clear and safe way for AI assistants to work with local services and APIs while keeping the user in control.

## What does this server do?

### Capabilities

- [x] Get current account information (`tool: tg_me`)
- [x] List dialogs with optional unread filter (`tool: tg_dialogs`)
- [x] Mark dialog as read (`tool: tg_read`)
- [x] Retrieve messages from specific dialog (`tool: tg_dialog`)
- [x] Send draft messages to any dialog (`tool: tg_send`)

### Prompt examples

Here are some example prompts you can use with AI assistants:

#### Message Management
- "Check for any unread important messages in my Telegram"
- "Summarize all my unread Telegram messages"
- "Read and analyze my unread messages, prepare draft responses where needed"
- "Check non-critical unread messages and give me a brief overview"

#### Organization
- "Analyze my Telegram dialogs and suggest a folder structure"
- "Help me categorize my Telegram chats by importance"
- "Find all work-related conversations and suggest how to organize them"

#### Communication
- "Monitor specific chat for updates about [topic]"
- "Draft a polite response to the last message in [chat]"
- "Check if there are any unanswered questions in my chats"

## Installation

### Homebrew

You can install a binary release on macOS/Linux using brew:

```bash
# Install
brew install chaindead/tap/telegram-mcp

# Update
brew upgrade chaindead/tap/telegram-mcp
```

### NPX

You can run the latest version directly using npx (supports macOS, Linux, and Windows):

```bash
npx -y @chaindead/telegram-mcp
```

When using NPX, modify the standard commands and configuration as follows:

- [Authentication command](#authorization) becomes:
```bash
npx -y @chaindead/telegram-mcp auth ...
```

- [Claude MCP server configuration](#client-configuration) becomes:
```json
{
  "mcpServers": {
    "telegram": {
      "command": "npx",
      "args": ["-y", "@chaindead/telegram-mcp"],
      "env": {
        "TG_APP_ID": "<your-api-id>",
        "TG_API_HASH": "<your-api-hash>"
      }
    }
  }
}
```

For complete setup instructions, see [Authorization](#authorization) and [Client Configuration](#client-configuration).

### From Releases

#### MacOS

<details>

> **Note:** The commands below install to `/usr/local/bin`. To install elsewhere, replace `/usr/local/bin` with your preferred directory in your PATH.

First, download the archive for your architecture:

```bash
# For Intel Mac (x86_64)
curl -L -o telegram-mcp.tar.gz https://github.com/chaindead/telegram-mcp/releases/latest/download/telegram-mcp_Darwin_x86_64.tar.gz

# For Apple Silicon (M1/M2)
curl -L -o telegram-mcp.tar.gz https://github.com/chaindead/telegram-mcp/releases/latest/download/telegram-mcp_Darwin_arm64.tar.gz
```

Then install the binary:

```bash
# Extract the binary
sudo tar xzf telegram-mcp.tar.gz -C /usr/local/bin

# Make it executable
sudo chmod +x /usr/local/bin/telegram-mcp

# Clean up
rm telegram-mcp.tar.gz
```
</details>

#### Linux
<details>

> **Note:** The commands below install to `/usr/local/bin`. To install elsewhere, replace `/usr/local/bin` with your preferred directory in your PATH.

First, download the archive for your architecture:

```bash
# For x86_64 (64-bit)
curl -L -o telegram-mcp.tar.gz https://github.com/chaindead/telegram-mcp/releases/latest/download/telegram-mcp_Linux_x86_64.tar.gz

# For ARM64
curl -L -o telegram-mcp.tar.gz https://github.com/chaindead/telegram-mcp/releases/latest/download/telegram-mcp_Linux_arm64.tar.gz
```

Then install the binary:

```bash
# Extract the binary
sudo tar xzf telegram-mcp.tar.gz -C /usr/local/bin

# Make it executable
sudo chmod +x /usr/local/bin/telegram-mcp

# Clean up
rm telegram-mcp.tar.gz
```
</details>

#### Windows

<details>

#### Windows
1. Download the latest release for your architecture:
   - [Windows x64](https://github.com/chaindead/telegram-mcp/releases/latest/download/telegram-mcp_Windows_x86_64.zip)
   - [Windows ARM64](https://github.com/chaindead/telegram-mcp/releases/latest/download/telegram-mcp_Windows_arm64.zip)
2. Extract the `.zip` file
3. Add the extracted directory to your PATH or move `telegram-mcp.exe` to a directory in your PATH
</details>

### From Source

Requirements:
- Go 1.24 or later
- GOBIN in PATH

```bash
go install github.com/chaindead/telegram-mcp@latest
```

## Configuration

### Authorization

Before you can use the server, you need to connect to the Telegram API.

1. Get the API ID and hash from [Telegram API](https://my.telegram.org/auth)
2. Run the following command:
   > __Note:__
   > If you have 2FA enabled: add --password <2fa_password>

   >  __Note:__
   > If you want to override existing session: add --new

   ```bash
   telegram-mcp auth --app-id <your-api-id> --api-hash <your-api-hash> --phone <your-phone-number>
   ```

   📩 Enter the code you received from Telegram to connect to the API.

3. Done! Please give this project a ⭐️ to support its development.

### Client Configuration

#### Claude Desktop

1. Open the Claude Desktop configuration file:
    - in MacOS, the configuration file is located at `~/Library/Application Support/Claude/claude_desktop_config.json`
    - in Windows, the configuration file is located at `%APPDATA%\Claude\claude_desktop_config.json`

   > __Note:__
   > You can also find claude_desktop_config.json inside the settings of Claude Desktop app

2. Add the server configuration:
   ```json
    {
      "mcpServers": {
        "telegram": {
          "command": "telegram-mcp",
          "env": {
            "TG_APP_ID": "<your-app-id>",
            "TG_API_HASH": "<your-api-hash>",
            "PATH": "<path_to_telegram-mcp_binary_dir>",
            "HOME": "<path_to_your_home_directory"
          }
        }
      }
    }
   ```

#### GitHub Copilot Desktop

1. Open the GitHub Copilot Desktop configuration file:
    - in MacOS, the configuration file is located at `~/.copilot/desktop/config.json`
    - in Windows, the configuration file is located at `%APPDATA%\GitHub\Copilot\hosts.json`

2. Add the server configuration:
   ```json
    {
      "mcpServers": {
        "telegram": {
          "command": "telegram-mcp",
          "env": {
            "TG_APP_ID": "<your-app-id>",
            "TG_API_HASH": "<your-api-hash>"
          }
        }
      }
    }
   ```

#### Cursor

1. Add the server configuration to your Cursor MCP settings:
   ```json
    {
      "mcpServers": {
        "telegram-mcp": {
          "command": "telegram-mcp",
          "env": {
            "TG_APP_ID": "<your-app-id>",
            "TG_API_HASH": "<your-api-hash>"
          }
        }
      }
    }
    ```

### JSON Schema Version

Some MCP clients (e.g. VS Code) do not support JSON Schema Draft 2020-12 and will reject tools that use it. You can override the JSON Schema version by setting the `--schema-version` flag or the `TG_SCHEMA_VERSION` environment variable.

Common values:
| Version | URL |
|---------|-----|
| Draft-07 (recommended for VS Code) | `https://json-schema.org/draft-07/schema#` |
| Draft 2020-12 (default) | `https://json-schema.org/draft/2020-12/schema` |

## Star History

<a href="https://www.star-history.com/#chaindead/telegram-mcp&Date">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=chaindead/telegram-mcp&type=Date&theme=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=chaindead/telegram-mcp&type=Date" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=chaindead/telegram-mcp&type=Date" />
 </picture>
</a>