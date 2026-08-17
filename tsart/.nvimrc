nnoremap <silent> <leader>tb :!make<CR>
nnoremap <silent> <leader>ai :Codex<CR>
nnoremap <silent> <leader>ti :!make install<CR>

" When editing this checkout, run the locally built host. This also avoids
" sandboxed Neovim sessions being unable to execute a binary under ~/.config.
let g:tsart_binary = expand('<sfile>:p:h') . '/bin/tsart'
execute 'source ' . fnameescape(expand('<sfile>:p:h') . '/register_plugin.vim')
