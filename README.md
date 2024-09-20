# GOCK3-LSP - Language Server for PDXScript in Crusader Kings 3

**GOCK3-LSP** is a Language Server Protocol (LSP) implementation for PDXScript files used in [Crusader Kings 3](https://www.crusaderkings.com/). It leverages the [GOCK3](https://github.com/unLomTrois/gock3) library to provide real-time feedback, code completion, and other language features in compatible editors.

## Features

- **Syntax Highlighting**: Enhanced readability with proper syntax coloring.
- **Real-time Linting**: Immediate detection of syntax and lexical errors.
- **Code Completion**: Intelligent suggestions based on context.
- **Symbol Navigation**: Easily jump to definitions and references.
- **Hover Information**: Inline documentation and tooltips.

## Table of Contents

- [Supported Editors](#supported-editors)
- [Contributing](#contributing)
- [Acknowledgments](#acknowledgments)
- [License](#license)

## TODO

The **GOCK3-LSP** project is currently under development. Below is a list of features and tasks to be implemented:

- [ ] **Initialize LSP Server with JSON-RPC 2.0**
- [ ] **Implement Text Document Synchronization**
  - [ ] Full synchronization
  - [ ] Incremental synchronization
- [ ] **Implement Diagnostics (Linting)**
  - [ ] Integrate with GOCK3 linter
  - [ ] Display syntax and lexical errors
- [ ] **Implement Code Completion**
  - [ ] Context-aware suggestions
  - [ ] Snippet support
- [ ] **Implement Hover Information**
  - [ ] Display documentation and type information
- [ ] **Implement Go to Definition**
  - [ ] Navigate to variable and function definitions
- [ ] **Implement Symbol Navigation**
  - [ ] Outline view support
- [ ] **Implement Formatting Support**
  - [ ] Code formatting based on predefined rules
- [ ] **Integrate with GOCK3 Library**
  - [ ] Utilize lexer, parser, and validator
- [ ] **Write Unit and Integration Tests**
  - [ ] Ensure reliability and stability
- [ ] **Set Up Continuous Integration (CI)**
  - [ ] Automated testing and builds
- [ ] **Create Comprehensive Documentation**
  - [ ] Usage guides
  - [ ] Configuration options
- [ ] **Release Initial Version**
  - [ ] Versioning
  - [ ] Distribution via GitHub Releases
- [ ] **Gather User Feedback and Iterate**
  - [ ] Implement user-requested features
  - [ ] Improve based on feedback


## Usage

The language server is designed to be launched by an LSP-compatible editor. Configure your editor to use `gock3-lsp` for PDXScript files.

## Supported Editors

- **Visual Studio Code**: Use the [GOCK3-VSCode Extension](https://github.com/unLomTrois/gock3-vscode).
- **Neovim**: Configure using `nvim-lspconfig` or similar plugins.

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

## Acknowledgments

- [GOCK3](https://github.com/unLomTrois/gock3) - The core library for parsing and validating PDXScript.
- [Language Server Protocol](https://microsoft.github.io/language-server-protocol/) - The protocol that enables rich language features across editors.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
