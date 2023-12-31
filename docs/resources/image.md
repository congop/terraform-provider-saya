---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "saya_image Resource - terraform-provider-saya"
subcategory: ""
description: |-
  Pull a saya Image to the host from the configure repository
---

# saya_image (Resource)

Pull a saya Image to the host from the configure repository

## Example Usage

```terraform
resource "saya_image" "test" {
  name      = "ubuntu:v1"
  img_type  = "ova"
  platform  = "linux/arm64"
  repo_type = "http"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `img_type` (String) image type, e.g. ova|vmdk, vhd, img, qcow2, etc.
- `name` (String) image name[:tag], e.r. appserver:v1

### Optional

- `hash` (String) hash to verify the pulled image, format <hash-type>:<hash-value>; md5:7287292, sha256:1234555
- `keep_locally` (Boolean) true to keep the pulled image in the local image store, false otherwise
- `platform` (String) image platform, e.g linux/amd64, linux/arm64/v7, defaults to host platform
- `repo_type` (String) the type of the remote repository; e.g.: http|s3

### Read-Only

- `id` (String) image identifier
- `sha256` (String) image sha256
