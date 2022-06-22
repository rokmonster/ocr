---
title: WSL
permalink: /issues/wsl
parent: Known Issues
---

# WSL

If you are using Windows Subsystem For Linux (WSL) you might have issues installing/running the tools.

Main issue is that WSL is based on older Linux distribution, so precompiled packages (deb/rpm) are linked with never version of libtesseract.

## Workarounds

* Install using [golang](../install/golang) method. (It should compile the binary against currently installed version of tesseract) 

**Workaround is not tested**