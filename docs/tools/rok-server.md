---
title: rok-server
nav_order: 2
permalink: /components/rok-server
parent: Components
---

# Server

`rok-server` enables WEB UI And easy access to data

## Features

* Image upload using web interface
* Defining data extraction zones using web interface
* Processing images (running jobs) (extracting data)

## Future Plans

* Pipelines (Combine data from multiple images/jobs)

## QuickStart (Ubuntu 22.04 / No TLS)

- Download either deb/rpm package from [Latest release](https://github.com/rokmonster/ocr/releases/latest/) 
- Install it with package manager. (`apt install -y ./package.deb`)
- Start with a simple `rok-server` & open [http://localhost:8080](http://localhost:8080) 

## QuickStart (Ubuntu 22.04 & TLS)

- Download either deb/rpm package from [Latest release](https://github.com/rokmonster/ocr/releases/latest/) 
- Install it with package manager. (`apt install -y ./package.deb`)
- Setup automatic startup & TLS
- Open your browser at https://${IP}.nip.io
  
```bash
# this requires domain name pointing to that server. If you only have IP, use ${IP}.nip.io
IP=$(curl -s https://ipinfo.io/ip)
rok-server -install -tls -domain ${IP}.nip.io -user $(whoami) | bash
echo "Open your browser at https://${IP}.nip.io"
```