# Terraform Provider for Site24x7

This is a [custom provider]( https://www.terraform.io/docs/extend/writing-custom-providers.html) 
for [Site24x7 API](https://www.site24x7.com/help/api/) 

`example.tf` file with the provider section:

```
provider "site24x7" {
  oauth_client_id = "your client id"
  oauth_client_secret = "your client secret"
  oauth_refresh_token = "your refresh token"
}
```

You can choose to specify the OAuth2 values as environment variables instead of including them in the provider section:

* `SITE24X7_CLIENT_ID`
* `SITE24X7_CLIENT_SECRET`
* `SITE24X7_REFRESH_TOKEN`

## Installation

```code
$ brew install terraform
$ go get -u github.com/sourcegraph/terraform-provider-site24x7 && cp $GOPATH/bin/terraform-provider-site24x7 .
$ echo 'providers { site24x7 = "terraform-provider-site24x7" }' >> ~/.terraformrc
```

### Obtaining new OAuth2 credentials

NOTE: This is only needed if the current credentials do not work (because scope changed for example).
 The current scope is `Site24x7.Admin.All`. 

If for some reason the OAuth2 credentials stored in `example.tf` don't work anymore or scope needs to change
do the following:

Follow [these instructions](https://www.site24x7.com/help/api/index.html#authentication) to obtain a client id,
 client secret and generate code token. 
 
Feed those into the script below:

```code
$ cd site24x7/oauth/cmd/site24x7-oauth
$ go run site24x7-oauth -clientId <someid> -clientSecret <somesecret> -generateCode <sometoken>
```

It will print to stdout the contents to be stored in `example.tf`.
