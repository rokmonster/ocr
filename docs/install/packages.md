---
layout: default
title: System Packages
nav_order: 1
permalink: /install/os-packages
parent: Installation
---

# Install using deb package

**This method of installing is the recommended for novice users, but only works on Linux VM's**

- Download either deb/rpm package from [Latest release](https://github.com/rokmonster/ocr/releases/latest/) 
- Install it with package manager. (`apt install -y ./package.deb`)

**Ubuntu/Debian**

```bash
curl -Lo /tmp/ocr.deb https://github.com/rokmonster/ocr/releases/latest/download/rokmonster-ocr_linux_amd64.deb
sudo apt install -y /tmp/ocr.deb
```


## Usage

Both packages deliver two binaries - `rok-scanner` & `rok-server`

- `rok-scanner` is used to scan a folder with images using one of templates;
- `rok-server` brings a web-based UI for easy use. This WebUI is reachable on [http://localhost:8080/](http://localhost:8080/)