---
layout: default
title: GoLang (MacOS)
nav_order: 3
permalink: /install/golang-mac-detailed
parent: Installation
---

# Install with GoLang (On A MAC / Should work on M1/M2 too)

```bash
# Install golang first:
curl -LO https://go.dev/dl/go1.19.2.darwin-amd64.pkg && open ./go1.19.2.darwin-amd64.pkg
# Go through install of go...

# next - install brew
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

#### in the end it will print some instructions to fix PATH variable. DO NOT SKIP THEM!!!

```bash
# install tesseract & opencv
brew install tesseract opencv
# You might need these
# export CPATH="/opt/homebrew/include"
# export LIBRARY_PATH="/opt/homebrew/lib"
```

### PLEASE CHECK FOR ERRORS!!!

```bash
Error: python@3.10: the bottle needs the Apple Command Line Tools to be installed.
  You can install them, if desired, with:
    xcode-select --install

If you're feeling brave, you can try to install from source with:
  brew install --build-from-source python@3.10
```

### Read every error. They will have suggestions how to fix them.

```bash
# finally, you should be able to download & compile rok-server or rok-remote'

go install github.com/rokmonster/ocr/cmd/rok-server@latest
go install github.com/rokmonster/ocr/cmd/rok-scanner@latest

```

The final binaries will be available at `~/go/bin`;

so run them like this:

```bash
~/go/bin/rok-server  # for server
~/go/bin/rok-scanner # for standalone scanner
```