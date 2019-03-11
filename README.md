# Terraform GReseller Provider

This is a terraform provider to add feature needed to act as a Google Cloud (or a GSuite in the future) reseller.
This module was heavily inspired by https://github.com/DeviaVir/terraform-provider-gsuite/

## Usage
Please see first installation to use the plugin.
If you already use GCP Terraform Provider, the most forward usage would be 
```
provider "greseller" { }

resource "greseller_cloud_billing_account" "a_billing_account" {
  display_name = "MyAccount"
  master_billing_account = "billingAccounts/MASTER_BILLING_ACCOUNT_ID"
}
```

## Authentication

There are two possible authentication mechanisms for using this provider.
Using a service account, or a personal account. The latter requires
user interaction, whereas a service account could be used in an automated
workflow.

```
provider "greseller" {
}
```


### Using a service account

Service accounts are great for automated workflows.

Add `credentials` when initializing the provider.
```
provider "gsuite" {
  credentials = "/full/path/service-account.json"
}
```

Credentials can also be provided via the following environment variables:
- GOOGLE_CREDENTIALS
- GOOGLE_CLOUD_KEYFILE_JSON
- GCLOUD_KEYFILE_JSON
- IMPERSONATED_USER_EMAIL

### Using a personal account

Use Google Cloud SDK login mechanism and then run:
```
$ gcloud auth application-default login
```

## Installation

1. Download the latest compiled binary from [GitHub releases](https://github.com/ekino/terraform-provider-greseller/releases).

1. Unzip/untar the archive.

1. Move it into `$HOME/.terraform.d/plugins`:

    ```sh
    $ mkdir -p $HOME/.terraform.d/plugins
    $ mv terraform-provider-greseller $HOME/.terraform.d/plugins/terraform-provider-greseller
    ```

1. Create your Terraform configurations as normal, and run `terraform init`:

    ```sh
    $ terraform init
    ```

    This will find the plugin locally.

## Available resources

### Sub billing account
This allows you to create a sub billing account as a google cloud reseller.
```
resource "greseller_cloud_billing_account" "a_billing_account" {
  display_name = "MyAccount"
  master_billing_account = "billingAccounts/MASTER_BILLING_ACCOUNT_ID"
}
```

Properties:
* display_name: Sub billing account display name
* master_billing_account: master billing account id in the format billingAccounts/MASTER_BILLING_ACCOUNT_ID

#### Caveats
Google currently don't allow to close / open a sub account through API.
A subaccount can not be deleted.
WARNING: When you remove the subaccount from terraform, it will juste remove the reference from terraform but won't close the subaccount nor delete it.

## Development

1. `cd` into `$HOME/.terraform.d/plugins/terraform-provider-greseller`

1. Run `make dep` to fetch the go vendor files

1. Make your changes

1. Run `make dev` and in your `terraform` directory, remove the current `.terraform` and re-run `terraform init`

1. Next time you run `terraform plan` it'll use your updated version