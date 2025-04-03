group "default" {
  targets = ["rok-server"]
}

target "_cross" {
  platforms = ["linux/amd64", "linux/arm64"]
  no-cache  = true
  pull      = true
}

target "rok-server" {
  inherits = ["_cross"]
  context = "./"
  tags     = [
    "ghcr.io/rokmonster/ocr:latest",
    "ghcr.io/rokmonster/ocr:${formatdate("YYYYMMDD", timestamp())}",
  ]
}
