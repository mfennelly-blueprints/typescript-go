if exists('g:loaded_tsart_config')
  finish
endif
let g:loaded_tsart_config = 1

" Start tsart as a remote plugin so it receives the live buffer, not just the
" last version written to disk. Put the cursor on a TypeScript symbol and use
" <leader>ta to see the AST node selected by tsart.
let s:tsart_binary = get(g:, 'tsart_binary', expand('~/.config/tsart/tsart'))
let s:tsart_log_file = get(g:, 'tsart_log_file', stdpath('log') . '/tsart.log')

" RPC owns stdout, but Neovim exposes a job's stderr separately.  Preserve it
" so a host startup failure or panic is not reduced to "Invalid channel".
function! s:TsartLog(lines) abort
  try
    call writefile(a:lines, s:tsart_log_file, 'a')
  catch
    echomsg 'tsart: unable to write log ' . s:tsart_log_file . ': ' . v:exception
  endtry
endfunction

function! s:OnTsartStderr(job_id, data, event) abort
  " A single empty item is the EOF notification, not a log entry.
  if a:data !=# ['']
    call s:TsartLog(a:data)
  endif
endfunction

function! s:OnTsartExit(job_id, code, event) abort
  call s:TsartLog([strftime('%Y-%m-%dT%H:%M:%S%z') . ' tsart host job ' . a:job_id . ' exited with status ' . a:code])
endfunction

function! s:RequireTsart(host) abort
  call s:TsartLog([strftime('%Y-%m-%dT%H:%M:%S%z') . ' starting tsart host: ' . s:tsart_binary . ' nvim'])
  let l:job = jobstart([s:tsart_binary, 'nvim'], {
  \ 'rpc': v:true,
  \ 'on_stderr': function('s:OnTsartStderr'),
  \ 'on_exit': function('s:OnTsartExit'),
  \ })
  if l:job <= 0
    call s:TsartLog([strftime('%Y-%m-%dT%H:%M:%S%z') . ' failed to start tsart host (jobstart returned ' . l:job . ')'])
  endif
  return l:job
endfunction

command! -bar TsartLog echo 'tsart log: ' . s:tsart_log_file

" Register the plugin.
call remote#host#Register('tsart', 'x', function('s:RequireTsart'))

let s:RemoteHostName = 'tsart'
let s:RemoteHostChannelIdentifier = '0'

" Register the commands for this plugin.
let s:ASTSelectCommandRPC = {
\ 'type': 'command',
\ 'name': 'TsartSelectAST',
\ 'sync': 1,
\ 'opts': {
\   'eval': '{''FileName'': expand(''%:p''), ''Text'': join(getline(1, ''$''), "\\n"), ''Line'': line(''.''), ''Column'': col(''.'') - 1}',
\ },
\}

let s:ASTHighlightCmd = {
\ 'type': 'command',
\ 'name': 'TsartHighlightAST',
\ 'sync': 1,
\ 'opts': {
\   'eval': '{''FileName'': expand(''%:p''), ''Text'': join(getline(1, ''$''), "\\n"), ''Line'': line(''.''), ''Column'': col(''.'') - 1}',
\ },
\}

let s:ASTHighlightClearCommand = {
\ 'type': 'command',
\ 'name': 'TsartClearASTHighlight',
\ 'sync': 1,
\ 'opts': {},
\}

let s:AnnotateSelectionCmd = {
\ 'type': 'command',
\ 'name': 'TsartAnnotateSelection',
\ 'sync': 1,
\ 'opts': {
\   'nargs': '+',
\   'range': '',
\   'eval': '{''Text'': join(getline(1, ''$''), "\\n"), ''StartLine'': line("''<"), ''StartColumn'': col("''<"), ''EndLine'': line("''>"), ''EndColumn'': col("''>"), ''Mode'': visualmode()}',
\ },
\}

let s:RPCCommands = [
\ s:ASTSelectCommandRPC,
\ s:ASTHighlightCmd,
\ s:ASTHighlightClearCommand,
\ s:AnnotateSelectionCmd,
\ ]

call remote#host#RegisterPlugin(s:RemoteHostName, s:RemoteHostChannelIdentifier, s:RPCCommands)
