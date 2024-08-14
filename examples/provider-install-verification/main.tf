terraform {
  required_providers {
    nkey = {
      source = "registry.terraform.io/holoplot/nkey"
    }
  }
}

provider "nkey" {
}

resource "nkey_nkey" "verify" {}