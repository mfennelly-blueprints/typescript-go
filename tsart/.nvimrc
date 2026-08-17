nnoremap <silent> <leader>tb :!make<CR>
nnoremap <silent> <leader>ai :Codex<CR>
nnoremap <silent> <leader>ti :!make install<CR>

" Start tsart as a remote plugin so it receives the live buffer, not just the
" last version written to disk. Put the cursor on a TypeScript symbol and use
" <leader>ta to see the AST node selected by tsart.
let s:tsart_binary = expand('~/.config/tsart/tsart')
function! s:RequireTsart(host) abort
  return jobstart([s:tsart_binary, 'nvim'], {'rpc': v:true})
endfunction

" register the plugin
call remote#host#Register('tsart', 'x', function('s:RequireTsart'))

let s:RemoteHostName = 'tsart'
let s:RemoteHostChannelIdentifier = '0'

" Register the commands for this plugin
let s:ASTSelectCommandRPC = {'type': 'command', 'name': 'TsartSelectAST', 'sync': 1,
         \  'opts': {'eval': '{''FileName'': expand(''%:p''), ''Text'': join(getline(1, ''$''), "\\n"), ''Line'': line(''.''), ''Column'': col(''.'') - 1}'}}
let s:ASTHighlightCmd = {'type': 'command', 'name': 'TsartHighlightAST', 'sync': 1,
         \  'opts': {'eval': '{''FileName'': expand(''%:p''), ''Text'': join(getline(1, ''$''), "\\n"), ''Line'': line(''.''), ''Column'': col(''.'') - 1}'}} 
let s:ASTHighlightClearCommand = {'type': 'command', 'name': 'TsartClearASTHighlight', 'sync': 1, 'opts': {}}
let s:AnnotateSelectionCmd = {'type': 'command', 'name': 'TsartAnnotateSelection', 'sync': 1,
         \  'opts': {'nargs': '+', 'range': '', 'eval': '{''Text'': join(getline(1, ''$''), "\\n"), ''StartLine'': line("''<"), ''StartColumn'': col("''<"), ''EndLine'': line("''>"), ''EndColumn'': col("''>"), ''Mode'': visualmode()}'}}

let s:RPCCommands = [
\ s:ASTSelectCommandRPC,
\ s:ASTHighlightCmd,
\ s:ASTHighlightClearCommand,
\ s:AnnotateSelectionCmd,
\ ]

call remote#host#RegisterPlugin(s:RemoteHostName, s:RemoteHostChannelIdentifier, s:RPCCommands)

nnoremap <silent> <leader>ta :TsartSelectAST<CR>
nnoremap <silent> <leader>th :TsartHighlightAST<CR>
nnoremap <silent> <leader>tH :TsartClearASTHighlight<CR>
