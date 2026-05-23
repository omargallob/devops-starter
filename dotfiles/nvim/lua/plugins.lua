-- Plugin specs for lazy.nvim

return {
  -- Colorscheme
  {
    "catppuccin/nvim",
    name = "catppuccin",
    priority = 1000,
    config = function()
      vim.cmd.colorscheme("catppuccin")
    end,
  },

  -- Treesitter
  {
    "nvim-treesitter/nvim-treesitter",
    build = ":TSUpdate",
    config = function()
      require("nvim-treesitter.configs").setup({
        ensure_installed = { "lua", "go", "python", "bash", "hcl", "yaml", "json", "toml", "dockerfile" },
        highlight = { enable = true },
      })
    end,
  },

  -- Telescope
  {
    "nvim-telescope/telescope.nvim",
    dependencies = { "nvim-lua/plenary.nvim" },
    keys = {
      { "<leader>ff", "<cmd>Telescope find_files<cr>" },
      { "<leader>fg", "<cmd>Telescope live_grep<cr>" },
      { "<leader>fb", "<cmd>Telescope buffers<cr>" },
    },
  },

  -- Git signs
  {
    "lewis6991/gitsigns.nvim",
    config = true,
  },

  -- LSP
  {
    "neovim/nvim-lspconfig",
  },

  -- Completion
  {
    "hrsh7th/nvim-cmp",
    dependencies = { "hrsh7th/cmp-nvim-lsp" },
    config = function()
      local cmp = require("cmp")
      cmp.setup({
        sources = { { name = "nvim_lsp" } },
        mapping = cmp.mapping.preset.insert({
          ["<CR>"] = cmp.mapping.confirm({ select = true }),
        }),
      })
    end,
  },

  -- File explorer
  {
    "stevearc/oil.nvim",
    config = true,
    keys = {
      { "-", "<cmd>Oil<cr>" },
    },
  },

  -- Which-key
  {
    "folke/which-key.nvim",
    event = "VeryLazy",
    config = true,
  },
}
