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