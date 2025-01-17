---
layout: "vault"
page_title: "Vault: vault_identity_entity_alias resource"
sidebar_current: "docs-vault-resource-identity-entity-alias"
description: |-
  Creates an Identity Entity Alias for Vault.
---

# vault\_identity\_entity\_alias

Creates an Identity Entity Alias for Vault. 

~> **Important** All data provided in the resource configuration will be
written in cleartext to state and plan files generated by Terraform, and
will appear in the console output when Terraform runs. Protect these
artifacts accordingly. See
[the main provider documentation](../index.html)
for more details.

## Example Usage

```hcl
resource "vault_identity_entity_alias" "test" {
  name            = "user_1"
  mount_accessor  = "token_1f2bd5"
  canonical_id    = "49877D63-07AD-4B85-BDA8-B61626C477E8"
}
```

## Argument Reference

The following arguments are supported:

* `namespace` - (Optional) The namespace to provision the resource in.
  The value should not contain leading or trailing forward slashes.
  The `namespace` is always relative to the provider's configured [namespace](../index.html#namespace).
   *Available only for Vault Enterprise*.

* `name` - (Required) Name of the alias. Name should be the identifier of the client in the authentication source. For example, if the alias belongs to userpass backend, the name should be a valid username within userpass backend. If alias belongs to GitHub, it should be the GitHub username.

* `mount_accessor` - (Required) Accessor of the mount to which the alias should belong to.

* `canonical_id` - (Required) Entity ID to which this alias belongs to.


## Attributes Reference

* `id` - ID of the entity alias.

## Import

Identity entity alias can be imported using the `id`, e.g.

```
$ terraform import vault_identity_entity_alias.test "3856fb4d-3c91-dcaf-2401-68f446796bfb"
```
