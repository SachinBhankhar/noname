# noname LSP

The noname language server provides diagnostics, hover, completion, and go-to-definition for `.nn` files.

## Building

```bash
go build -o ~/.local/bin/noname-lsp ./lsp/
```

Make sure `~/.local/bin` is on your `$PATH`.

---

## Neovim

No plugins required — Neovim has a built-in LSP client.

Add the following to your Neovim config (e.g. `~/.config/nvim/lua/noname.lua`) and require it from your `init.lua`:

```lua
vim.filetype.add({ extension = { nn = "noname" } })

vim.api.nvim_create_autocmd("FileType", {
  pattern = "noname",
  callback = function()
    vim.lsp.start({
      name = "noname-lsp",
      cmd = { "noname-lsp" },
      root_dir = vim.fs.dirname(
        vim.fs.find({ "go.mod", ".git" }, { upward = true })[1]
      ),
      capabilities = vim.lsp.protocol.make_client_capabilities(),
    })
  end,
})
```

For syntax highlighting, create `~/.config/nvim/syntax/noname.vim` — see the repo's `editor/neovim/` folder for the ready-made file.

### Keymaps (already standard in most configs)

| Key           | Action              |
|---------------|---------------------|
| `K`           | Hover               |
| `gd`          | Go to definition    |
| `[d` / `]d`   | Previous/next error |
| `<leader>vd`  | Show error details  |
| `Ctrl+Space`  | Trigger completion  |

---

## VS Code

VS Code requires a small extension to register the language server.

### 1. Install dependencies

```bash
npm install -g yo generator-code vsce
```

### 2. Scaffold the extension

```bash
yo code
# Choose: New Language Support
# Language id: noname
# File extensions: .nn
# Choose: No (don't add a grammar)
```

### 3. Replace the generated `extension.js`

```js
const { LanguageClient, TransportKind } = require('vscode-languageclient/node');

let client;

function activate(context) {
  client = new LanguageClient(
    'noname-lsp',
    'noname Language Server',
    {
      command: 'noname-lsp',
      transport: TransportKind.stdio,
    },
    {
      documentSelector: [{ scheme: 'file', language: 'noname' }],
    }
  );
  client.start();
}

function deactivate() {
  if (client) return client.stop();
}

module.exports = { activate, deactivate };
```

### 4. Register the language in `package.json`

```json
"contributes": {
  "languages": [{
    "id": "noname",
    "extensions": [".nn"]
  }]
}
```

### 5. Install the extension locally

```bash
vsce package
code --install-extension noname-lsp-0.0.1.vsix
```

---

## Zed

Zed supports LSP servers through its extension system.

### 1. Create the extension directory

```bash
mkdir -p ~/.config/zed/extensions/noname/languages/noname
```

### 2. Create `extension.toml`

```toml
id = "noname"
name = "noname"
version = "0.0.1"
description = "noname language support"
authors = ["you"]

[language_servers.noname-lsp]
name = "noname LSP"
language = "noname"
```

### 3. Create `languages/noname/config.toml`

```toml
name = "noname"
path_suffixes = ["nn"]
line_comment = "//"

[brackets]
  [[brackets.pair]]
  start = "{"
  end = "}"
  [[brackets.pair]]
  start = "("
  end = ")"
```

### 4. Create `languages/noname/highlights.scm`

Zed uses Tree-sitter for highlighting. For now add a placeholder — full Tree-sitter grammar support can be added later:

```scheme
; placeholder
```

### 5. Register the language server in Zed settings

Open Zed settings (`Cmd+,`) and add:

```json
{
  "lsp": {
    "noname-lsp": {
      "binary": {
        "path": "/home/YOU/.local/bin/noname-lsp"
      }
    }
  }
}
```

Replace `/home/YOU/` with your actual home directory path.

---

## Verifying the server is running

Create a test file:

```bash
cat > test.nn << 'EOF'
let x = 10;
let name = "alice";
print(name);
print(unknownVar);
EOF
```

Open it in your editor — you should see:
- `unknownVar` underlined in red with `identifier not found: unknownVar`
- Hover over `name` shows `string "alice"`
- Hover over `print` shows the builtin description
- Completion list appears with keywords, builtins, `x`, and `name`
