# Waypoint Secrets Engine for HashiCorp Vault

The waypoint secrets engine generates user tokens dynamically for a Waypoint server. This means that services that need to access a Waypoint server no longer need to hardcode tokens.

Vault makes use both of its own internal revocation system to delete waypoint users when generating waypoint credentials to ensure that tokens become invalid within a reasonable time of the lease expiring.

## Setup

Most secrets engines must be configured in advance before they can perform their functions. These steps are usually completed by an operator or configuration management tool.


1. Enable secrets engine:


```shell
vault secrets enable waypoint
```

By default, the secrets engine will mount at the name of the engine. To enable the secrets engine at a different path, use the -path argument.


2. Configure the credentials that Vault uses to communicate with waypoint to generate credentials:
```shell
vault write waypoint/config \
  addr=localhost:9701 \
  token=${WAYPOINT_TOKEN}
```

3. Configure a role that sets how long a token will be valid for:

```shell
vault write waypoint/role/my-role \
  ttl=180 \
  max_ttl=360 
```

By writing to the roles/my-role path we are defining the my-role role. 

## Usage

After the secrets engine is configured and a user/machine has a Vault token with the proper permission, it can generate credentials.

1. Generate a new credential by reading from the /creds endpoint with the name of the role:
```shell
vault read waypoint/creds/my-role
```

## API

### Setup

1. Enable secrets engine

Sample request

```shell
curl \
    -X POST \
    --header "X-Vault-Token: ..." \
    http://127.0.0.1:8200/v1/sys/mounts
```

Sample payload

```json
{
    "type": "waypoint"
}
```

2. Configure the credentials that Vault uses to communicate with waypoint to generate credentials:

Sample request
```shell
curl \
    -X POST \
    --header "X-Vault-Token: ..." \
    http://127.0.0.1:8200/v1/waypoint/config
```

Sample payload
```json
{
  "addr": "localhost:9701",
  "token": "insert waypoint token here"
}
```

3. Configure a role that maps a name in Vault to a waypoint scope and roles:

Sample request
```shell
curl \
    -X POST \
    --header "X-Vault-Token: ..." \
    http://127.0.0.1:8200/v1/waypoint/role/my-role
```

Sample payload
```json
{
    "ttl": 180,
    "max_ttl": 360
}
```

### Usage

1. Generate a new credential by reading from the /creds endpoint with the name of the role:

Sample request
```shell
curl \
    -X GET \
    --header "X-Vault-Token: ..." \
    http://127.0.0.1:8200/v1/waypoint/creds/my-role
```

Sample response
```json
{
    "request_id": "ed281bc6-182d-a15e-d700-8c2e64897010",
    "lease_id": "waypoint/creds/my-role/pH9CfQcAmE9va6CwQKOEPBsx",
    "renewable": true,
    "lease_duration": 180,
    "data": {
        "token": "BCkP8cw7qjrzhTt46...",
        "user_id": "01G1Y870WBTWR9JRTEGSQED6WZ"
    },
    "wrap_info": null,
    "warnings": null,
    "auth": null
}
```

## Terraform

### Setup

1. Enable secrets engine:

```hcl
resource "vault_mount" "waypoint" {
  path        = "waypoint"
  type        = "waypoint"
  description = "This is the waypoint secrets engine"
}
```

2. Configure the credentials that Vault uses to communicate with waypoint to generate credentials:

```hcl
resource "vault_generic_endpoint" "waypoint_config" {
  depends_on           = [
    vault_mount.waypoint
  ]
  
  path                 = "waypoint/config"
  ignore_absent_fields = true

  data_json = <<EOT
{
  "addr": "localhost:9701",
  "token": "..."
}
EOT
}

```

3. Configure a role that maps a name in Vault to a waypoint scope and roles:

```hcl
resource "vault_generic_endpoint" "waypoint_role" {
  depends_on           = [
    vault_mount.waypoint
  ]
  
  path                 = "waypoint/role/my-role"
  ignore_absent_fields = true

  data_json = <<EOT
{
    "ttl": 180,
    "max_ttl": 360
}
EOT
}
```

## Usage

1. Generate a new credential by reading from the /creds endpoint with the name of the role:

```hcl
data "vault_generic_secret" "waypoint_creds" {
  path = "waypoint/creds/my-role"
}

output "creds" {
  value     = data.vault_generic_secret.waypoint_creds.data
  sensitive = true
}
```

2. Read the output from Terraform's state file:

```shell
terraform output creds
```

Example response:

```
tomap({
  "token" = "BCkP8cw7qjrzhTt46..."
  "user_id" = "u_TxJs1IabfY"
})
```

## License

Licensed under the Apache License, Version 2.0 (the "License").

You may obtain a copy of the License at apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" basis, without WARRANTIES or conditions of any kind, either express or implied.

See the License for the specific language governing permissions and limitations under the License.