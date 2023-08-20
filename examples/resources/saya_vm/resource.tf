resource "saya_vm" "test" {
  name         = "test1vm"
  image        = "linux/amd64:webserver:v1:qcow2"
  compute_type = "qemu"
  state        = "running"
}