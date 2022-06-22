---
title: GoLang
nav_order: 3
permalink: /install/golang
parent: Installation
---

# Install with GoLang

This method of install is more flexible & allows you to run it on any machine (Linux/Mac/Windows)

In order for it to work you need to have [Go](https://go.dev/dl/) installed. It also requires you to have required libs (libtesseract) available on the system, as this method is actually compiling a binary from source.

This method will work fine on different architectures (arm64) too. So you can use it on Raspberry PI.

### ROK Server

```bash
go install github.com/xor22h/rok-monster-ocr-golang/cmd/rok-server@latest
$GOBIN/rok-server
```

### ROK Scanner

```bash
go install github.com/xor22h/rok-monster-ocr-golang/cmd/rok-scanner@latest
$GOBIN/rok-scanner
```