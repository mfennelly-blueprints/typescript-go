# tsart - Automated Refactoring Tools, for Typescript 7.0.2

- [Concepts](#concepts)
- [Setup Manual](#setup-manual)
- [User Manual](#user-manual)

----------
# Concepts
## Syntax Trees - Refactoring and Highlighting

The editor is a fork of neovim, which uses the neovim remote 
plugin interface to create abstract syntax tree based editing for users and 
agents.
Typically, editing is separated by revision/filesystem structure, rather than
the ontological contents of the codebase. This allows you to identify verticals 
for your agents based on where your code will be linked, rather than where it is
located on the file system.

So this is a different paradigm of editing and changing
code - semantic traversal rather than filesystem traversal.

## Creating a Refactor Decision Graph

This is a .jsonnet file that you can use to describe how to
perform refactors on TypeScript code. It matches a deviation
from some pattern, and plans a refactor to reach
the result.

----------
# Setup Manual

1. Build the remote plugin binary `make`
2. Use the ./vim/register_plugin.vim script by running `make install`.

Commands:

`TsartSelectAST`: Select a reference to an abstract syntax tree node.
`TsartHighlightAST`: Highlight an AST Node.
`TsartClearASTHighlight`: Clear a highlight on an AST node.
`TsartAnnotateSelection`: Annotate an arbitrary selection.

Common operations:

`TsartFunctionLZero`: 

Highlights the immediate `ast.Node` elements for symbols of type Kind that are 
direct children of the file including `KindParameter`, `KindVariableDeclaration`,
`KindReturnStatement`, `KindBindingElement`


----------
# User Manual

## Working With TypeScript
### Immediately available paradigms

#### Functional TypeScript

This expects a specific Syntax Tree shape.
Namely that your source files consist of:

1. Functions
2. Modifiers
3. Interfaces
4. Types

Your codebase is organised by the compiler in a tree.
When you are editing your codebase, you are editing a file.
The TypeScript compiler parses this as an `ast.SourceFile`.

An `ast.SourceFile` is a list of `ast.Statement` nodes, and 
an `ast.EndOfFile`.

In Functional TypeScript, we want the graph to look like

**SourceFile Structure:**
```txt
-> ast.Node {Kind: SourceFile}
   Statements: {
     // import declarations
     -> ast.Node {Kind: ImportDeclaration}
     // function declarations, types & interfaces with exports as modifiers
     -> ast.Node {Kind: FunctionDeclaration}
   }
```

**Function Structure:**
Rather than having functions with nested node definitions,
we want the functions to be extracted hierarchically.

1. All functions must have a FunctionBody.
2. No ambient declarations.
3. No generator functions.
4. Must have a type.

**Function Body**:

1. `KindVariableDeclaration`

All binding names must be identifiers.

All identifiers must have an associated type.

The initializer must be a simple expression, or an identifier. It can not be
things like:

Arrow Functions.
Complex Conditionals.

2. `IfStatement`: expression, thenStatement node must be a block, and complexity of that block.

3. `KindReturnStatement` must return an identifier. 

**Import statements:**

Import Phase modifiers (type, defer)
Namespaced imports must have names.
Named imports.
No module specifiers - use namespaces.

[NOTE]
We want to minimize the variance and narrow the scope of the structure of
ModuleSpecifier.

**Packages, modules and namespaces:**

packages:
I would prefer if these were version specified in source.

i.e. instead of...
```ts
import packageName from "packageName"
```

..you have:
```ts
import packageName from "packageName/v3.2"
```

Namespaces:
I like that these explicitly define the verticals of
what specific code needs. The namespace abstraction
in TypeScript/JavaScript is quite leaky already however.

### More on the TypeScript Syntax Tree, SyntaxKinds

A language is an alphabet and a grammar.
Tokens are also a concept that form the metasyntax.

TypeScript breaks this up further into "SyntaxKinds", bases, unions, nodes and bases. 

bases -> ...(35 variants)

nodes (extend bases) -> aliases (73), definitions (192), listAliases (24)

kinds (node's discriminant) -> elements (446), markers (138), aliases (221)

### Node Kind source code

In TypeScript 7.0.2 there are a load of different [kinds][kind-generated]. These come in different taxonomic classes:

- [Psuedo-Literals][pseudo-literals]
- [Punctuation][punctuation]
- [Assignments][assignments]
- [Identifiers/Private Identifiers][identifiers]
- [Reserved Keywords][reserved-keywords]
- [Strict Mode Reserved Keywords][strict-mode-reserved-keywords]
- [Contextual Keywords][contextual-keywords]

Tree Node Related Kinds:

- [Names][names]
- [Signature Elements][signature-elements]
- [TypeMember][type-members]
- [Types][types]
- [Binding Patterns][binding-patterns]
- [Expressions][expressions]
- [Element][elements]

These are produced by the compiler scanners (of which there are multiple).

[kind-generated]: https://github.com/microsoft/typescript-go/blob/typescript/v7.0.2/internal/ast/kind_generated.go
[pseudo-literals]: https://github.com/microsoft/typescript-go/blob/typescript/v7.0.2/internal/ast/kind_generated.go#L27-L29
[punctuation]: https://github.com/microsoft/typescript-go/blob/typescript/v7.0.2/internal/ast/kind_generated.go#L31-L73
[assignments]: https://github.com/microsoft/typescript-go/blob/typescript/v7.0.2/internal/ast/kind_generated.go#L79-L94
[identifiers]: https://github.com/microsoft/typescript-go/blob/typescript/v7.0.2/internal/ast/kind_generated.go#L96-L98
[reserved-keywords]: https://github.com/microsoft/typescript-go/blob/typescript/v7.0.2/internal/ast/kind_generated.go#L100-L135
[strict-mode-reserved-keywords]: https://github.com/microsoft/typescript-go/blob/typescript/v7.0.2/internal/ast/kind_generated.go#L137-L145
[contextual-keywords]: https://github.com/microsoft/typescript-go/blob/typescript/v7.0.2/internal/ast/kind_generated.go#L147-L186
[names]: https://github.com/microsoft/typescript-go/blob/typescript/v7.0.2/internal/ast/kind_generated.go#L189-L190
[signature-elements]: https://github.com/microsoft/typescript-go/blob/typescript/v7.0.2/internal/ast/kind_generated.go#L192-L194
[type-members]: https://github.com/microsoft/typescript-go/blob/typescript/v7.0.2/internal/ast/kind_generated.go#L196-L206
[types]: https://github.com/microsoft/typescript-go/blob/typescript/v7.0.2/internal/ast/kind_generated.go#L208-L231
[binding-patterns]: https://github.com/microsoft/typescript-go/blob/typescript/v7.0.2/internal/ast/kind_generated.go#L233-L235
[expressions]: https://github.com/microsoft/typescript-go/blob/typescript/v7.0.2/internal/ast/kind_generated.go#L237-L266
[elements]: https://github.com/microsoft/typescript-go/blob/typescript/v7.0.2/internal/ast/kind_generated.go#L271-L312
