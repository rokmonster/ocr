---
title: System Packages
nav_order: 1
permalink: /install/os-packages
parent: Installation
---

# Install using deb/rpm package

**This method of installing is the recommended for novice users, but only works on Linux VM's**

- Download either deb/rpm package from [Latest release](https://github.com/xor22h/rok-monster-ocr-golang/releases/latest/) 
- Install it with package manager. (`apt install -y ./package.deb` or `yum install -y ./package.rpm`)

**Ubuntu/Debian**

```bash
curl -Lo /tmp/rok-monster.deb https://github.com/xor22h/rok-monster-ocr-golang/releases/latest/download/rok-monster-ocr-golang.deb
sudo apt install -y /tmp/rok-monster.deb
```

**Centos/Redhat**

```bash
sudo yum install -y https://github.com/xor22h/rok-monster-ocr-golang/releases/latest/download/rok-monster-ocr-golang.rpm
```

## Usage

Both packages deliver two binaries - `rok-scanner` & `rok-server`

- `rok-scanner` is used to scan a folder with images using one of templates;
- `rok-server` brings a web-based UI for easy use. This WebUI is reachable on [http://localhost:8080/](http://localhost:8080/)