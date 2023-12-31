---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "saya Provider"
subcategory: ""
description: |-
  
---

# saya Provider



## Example Usage

```terraform
provider "saya" {
  version = "0.0.1"
  exe     = "saya"
  forge   = "/tmp/tempforge"

  http_repo = {
    url       = "http://localhost:10099"
    base_path = ""
  }

  s3_repo = {
    bucket   = "repos.example.com"
    base_key = "public"
    region   = "us-east-1"
    credentials = {
      access_key_id     = "xxxxxxxxxxxxxxxx"
      secret_access_key = "XXXXXXXXXXXX"
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `config` (String) saya yaml config path
- `exe` (String) saya exe command or path; defaults to saya
- `forge` (String) forge location
- `http_repo` (Attributes) (see [below for nested schema](#nestedatt--http_repo))
- `license_key` (String, Sensitive) license key value or file
- `s3_repo` (Attributes) (see [below for nested schema](#nestedatt--s3_repo))

<a id="nestedatt--http_repo"></a>
### Nested Schema for `http_repo`

Required:

- `url` (String) the url  of the remote repository

Optional:

- `base_path` (String) the base path  of the remote repository
- `basic_auth` (Attributes) (see [below for nested schema](#nestedatt--http_repo--basic_auth))
- `upload_strategy` (String) upload strategy

<a id="nestedatt--http_repo--basic_auth"></a>
### Nested Schema for `http_repo.basic_auth`

Required:

- `password` (String, Sensitive) the password
- `username` (String) the username to authenticate with



<a id="nestedatt--s3_repo"></a>
### Nested Schema for `s3_repo`

Required:

- `bucket` (String) the s3 bucket of the remote repository

Optional:

- `base_key` (String) the base key of the remote repository
- `credentials` (Attributes) (see [below for nested schema](#nestedatt--s3_repo--credentials))
- `ep_url_s3` (String) the endpoint url for the s3 service
- `region` (String) the region of the s3 service
- `use_path_style` (Boolean) true to allows the client to use path-style addressing, i.e., https://s3.amazonaws.com/BUCKET/KEY

<a id="nestedatt--s3_repo--credentials"></a>
### Nested Schema for `s3_repo.credentials`

Optional:

- `access_key_id` (String) aws access id
- `can_expire` (String) if the credential can expire
- `expires` (String) the time the credential will expire
- `secret_access_key` (String, Sensitive) aws secret access key
- `session_token` (String, Sensitive) the aws session token
- `source` (String) the source of the credential
