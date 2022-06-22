---
title: rok-server
nav_order: 2
permalink: /components/rok-server
parent: Components
---

## Server

`rok-server` enables WEB UI And easy access to data

## Features

* Image upload using web interface
* Defining data extraction zones using web interface
* Processing images (running jobs) (extracting data)

## Future Plans

* Pipelines (Combine data from multiple images/jobs)

## QuickStart (Linux VM / Local PC / No TLS)

- Download either deb/rpm package from [Latest release](https://github.com/xor22h/rok-monster-ocr-golang/releases/latest/) 
- Install it with package manager. (`apt install -y ./package.deb` or `yum install -y ./package.rpm`)
- Start with a simple `rok-server` & open [http://localhost:8080](http://localhost:8080) 

## QuickStart (Linux VM & TLS)

- Download either deb/rpm package from [Latest release](https://github.com/xor22h/rok-monster-ocr-golang/releases/latest/) 
- Install it with package manager. (`apt install -y ./package.deb` or `yum install -y ./package.rpm`)
- Setup automatic startup & TLS
- Open your browser at https://${IP}.nip.io
  
```bash
# this requires domain name pointing to that server. If you only have IP, use ${IP}.nip.io
IP=$(curl -s https://ipinfo.io/ip)
rok-server -install -tls -domain ${IP}.nip.io -user $(whoami) | bash
echo "Open your browser at https://${IP}.nip.io"
```