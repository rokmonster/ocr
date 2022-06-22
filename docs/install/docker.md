---
title: Docker
nav_order: 2
permalink: /install/docker
parent: Installation
---

# Install with Docker

If you are familiar with [Docker](https://www.docker.com/products/docker-desktop/), running `rok-server` is very easy & simple using only few commands.

**This method works on Windows/Linux/MacOS**

```bash
docker pull ghcr.io/xor22h/rok-monster-ocr-golang:latest
docker run -d -p8080:8080 ghcr.io/xor22h/rok-monster-ocr-golang:latest
```

Just open  [http://localhost:8080](http://localhost:8080) and enjoy.