# ROK Monster OCR (GoLang)

[![Discord](https://img.shields.io/discord/955712057766985789)](https://discord.gg/3YS3xszTXe) 
[![License: MIT](https://img.shields.io/github/license/xor22h/rok-monster-ocr-golang)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/xor22h/rok-monster-ocr-golang?style=flat-square)](https://goreportcard.com/report/github.com/xor22h/rok-monster-ocr-golang)


---

👋 An idea for this project came from [ROK Monster OCR Tools](https://github.com/carmelosantana/rok-monster-ocr).

---

👋 Join our [Discord](https://discord.gg/drhxwVQ) for help getting started or show off your results!

---

## Kingdom Statistics

Command line tools to help collect player statistics from [Rise of Kingdoms](https://rok.lilithgames.com/en). By analyzing screenshots we can extract various data points such as governor power, deaths, kills and more. This can help with various kingdom statistics or fairly distributing [KvK](https://rok.guide/the-lost-kingdom-kvk/) rewards.

![Sample](./media/sample.png)

[![asciicast](https://asciinema.org/a/gYerprrrw0DVOXZbitOfHrPqg.svg)](https://asciinema.org/a/gYerprrrw0DVOXZbitOfHrPqg)

## Features

- Character recognition by [Tesseract](https://github.com/tesseract-ocr/tesseract)
- Easy install with package managers `apt-get` / `yum`
- Fast hash based image comparison
- Automated pick of best-match template (based on first image in media directory)
- Easy to use WebUI. Just open [localhost:8080](http://localhost:8080/), upload files, and get results directly in your browser.
- Automatic download/update of Tesseract data files.

## Future Plans

- Ability to use multiple templates in single run
- Discord BOT mode. (Process each image sent to a specific discord channel)
- Automate screnshot collection using ADB & Memu/LDPLay/real android device

## Limitations

- English language is preferred as coordinate information lines up most accurately with English.
- No way to merge user information from different screens.
- Best template is detected automatically, but same template is used for all files in media directory.
- Requires a template defined for proper device (resolution/acpect-ratio/language)
- Limited number of predefined templates

## Getting started

[Read the docs here](https://xor22h.github.io/rok-monster-ocr-golang/)

## Community

Have a question, an idea, or need help getting started? Checkout our [Discord](https://discord.gg/drhxwVQ)!
