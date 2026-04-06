-- ccl-lsp configuration for nvim-lspconfig
-- Copy the contents of this file into your Neovim config (e.g., after/plugin/lsp.lua)
-- or require it from your lazy.nvim setup.

local lspconfig = require("lspconfig")
local configs = require("lspconfig.configs")

if not configs.ccl_lsp then
  configs.ccl_lsp = {
    default_config = {
      cmd = { "ccl-lsp" },
      filetypes = { "ccl" },
      root_dir = lspconfig.util.root_pattern(".git", ".ccolonfmt") or vim.fn.getcwd,
      single_file_support = true,
      settings = {},
    },
    docs = {
      description = "Language server for CColon (.ccl)",
    },
  }
end

lspconfig.ccl_lsp.setup({
  on_attach = function(_, bufnr)
    local opts = { buffer = bufnr, noremap = true, silent = true }
    vim.keymap.set("n", "gd",         vim.lsp.buf.definition,     opts)
    vim.keymap.set("n", "K",          vim.lsp.buf.hover,          opts)
    vim.keymap.set("n", "<leader>f",  vim.lsp.buf.format,         opts)
    vim.keymap.set("n", "<leader>ca", vim.lsp.buf.code_action,    opts)
  end,
  capabilities = require("cmp_nvim_lsp").default_capabilities(),
})
