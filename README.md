# ccl-lsp

Language Server Protocol implementation for [CColon](https://github.com/TRC-Loop/ccolon) — a bytecode-compiled programming language. Built with Go, using CColon's own lexer, parser, and formatter in-process.

**Language docs & reference:** [ccolon.arne.sh](https://ccolon.arne.sh)

## Features

| | |
|---|---|
| Diagnostics | Syntax and parse errors shown inline |
| Completion | Keywords, types, variables, functions, classes, stdlib modules, dot-completion |
| Hover | Signatures for symbols, docs for keywords and stdlib methods |
| Go to definition | Jump to declaration of any top-level symbol |
| Formatting | Full document formatting, respects `.ccolonfmt` config |
| Semantic tokens | Rich token classification for editors that support it |
| Syntax highlighting | Fallback `syntax/ccl.vim` for Vim/Neovim |

## Install

Requires Go 1.23+.

```sh
go install github.com/TRC-Loop/ccl-lsp@latest
```

Make sure `$(go env GOPATH)/bin` is in your `PATH`.

## Neovim setup (lazy.nvim + nvim-lspconfig)

The server is not yet in the Mason registry. Register it manually:

```lua
local lspconfig = require("lspconfig")
local configs = require("lspconfig.configs")

if not configs.ccl_lsp then
  configs.ccl_lsp = {
    default_config = {
      cmd = { "ccl-lsp" },
      filetypes = { "ccl" },
      root_dir = lspconfig.util.root_pattern(".git", ".ccolonfmt"),
      single_file_support = true,
    },
  }
end

lspconfig.ccl_lsp.setup({
  on_attach = function(_, bufnr)
    local opts = { buffer = bufnr, noremap = true, silent = true }
    vim.keymap.set("n", "gd",        vim.lsp.buf.definition,  opts)
    vim.keymap.set("n", "K",         vim.lsp.buf.hover,       opts)
    vim.keymap.set("n", "<leader>f", vim.lsp.buf.format,      opts)
  end,
})
```

### Filetype detection and syntax highlighting

Copy the vim files from `editor/` into your Neovim config:

```sh
cp editor/ftdetect/ccl.vim ~/.config/nvim/ftdetect/ccl.vim
cp editor/syntax/ccl.vim   ~/.config/nvim/syntax/ccl.vim
```

Or if you manage it as a plugin (e.g. with lazy.nvim's `dir` option pointing at this repo), Neovim will pick them up automatically.

## Formatter config

The formatter reads `.ccolonfmt` from the file's directory up to the filesystem root. Example:

```ini
indent_size = 4
use_tabs    = false
max_width   = 100
```

## License

MIT
