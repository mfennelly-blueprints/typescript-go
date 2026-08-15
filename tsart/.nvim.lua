vim.keymap.set("n", "<leader>tb", "<cmd>!make<CR>", {
   desc = "Build tsart",
})

vim.keymap.set("n", "<leader>ai", "<cmd>Codex<CR>", {
   desc = "Open Codex",
})

vim.keymap.set("n", "<leader>ti", "<cmd>!make install<CR>", {
   desc = "Install tsart",
})

-- Start tsart as a remote plugin so it receives the live buffer, not just the
-- last version written to disk. Put the cursor on a TypeScript symbol and use
-- <leader>ta to see the AST node selected by tsart.
vim.cmd([[ 
let s:tsart_binary = expand('~/.config/tsart/tsart')
function! s:RequireTsart(host) abort
  return jobstart([s:tsart_binary, 'nvim'], {'rpc': v:true})
endfunction
call remote#host#Register('tsart', 'x', function('s:RequireTsart'))
call remote#host#RegisterPlugin('tsart', '0', [
  \ {'type': 'command', 'name': 'TsartSelectAST', 'sync': 1,
  \  'opts': {'eval': '{''FileName'': expand(''%:p''), ''Text'': join(getline(1, ''$''), "\\n"), ''Line'': line(''.''), ''Column'': col(''.'') - 1}'}},
  \ {'type': 'command', 'name': 'TsartHighlightAST', 'sync': 1,
  \  'opts': {'eval': '{''FileName'': expand(''%:p''), ''Text'': join(getline(1, ''$''), "\\n"), ''Line'': line(''.''), ''Column'': col(''.'') - 1}'}},
  \ {'type': 'command', 'name': 'TsartClearASTHighlight', 'sync': 1, 'opts': {}},
  \ ])
]])

vim.keymap.set("n", "<leader>ta", "<cmd>TsartSelectAST<CR>", {
   desc = "Select cursor symbol in tsart AST",
})

vim.keymap.set("n", "<leader>th", "<cmd>TsartHighlightAST<CR>", {
   desc = "Highlight cursor token in tsart AST",
})

vim.keymap.set("n", "<leader>tH", "<cmd>TsartClearASTHighlight<CR>", {
   desc = "Clear tsart AST highlight",
})
