
data "saya_image" "readtest" {
  name     = "ubuntu:v_data_read_1"
  img_type = "ova"
  platform = "linux/arm64"
  filters = {
    label = ["audience=tester"]
  }
}
