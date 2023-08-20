
resource "saya_image" "test" {
  name      = "ubuntu:v1"
  img_type  = "ova"
  platform  = "linux/arm64"
  repo_type = "http"
}