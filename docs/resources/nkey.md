---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "nkey_nkey Resource - nkey"
subcategory: ""
description: |-
  An nkey is an ed25519 key pair formatted for use with NATS.
---

# nkey_nkey (Resource)

An nkey is an ed25519 key pair formatted for use with NATS.



<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `private_key` (String, Sensitive) Private key of the nkey to be given to the client for authentication
- `public_key` (String) Public key of the nkey to be given in config to the nats server
