---
title: Installation
nav_order: 1
permalink: /install
---

## Installation

- Download either deb/rpm package from [Latest release](https://github.com/xor22h/rok-monster-ocr-golang/releases/latest/) 
- Install it with package manager. (`apt install -y ./package.deb` or `yum install -y ./package.rpm`)

Ubuntu/Debian

```bash
curl -Lo /tmp/rok-monster.deb https://github.com/xor22h/rok-monster-ocr-golang/releases/latest/download/rok-monster-ocr-golang.deb
sudo apt install -y /tmp/rok-monster.deb
rok-scanner -help
```

Centos/Redhat

```bash
sudo yum install -y https://github.com/xor22h/rok-monster-ocr-golang/releases/latest/download/rok-monster-ocr-golang.rpm
rok-scanner -help
```