# vault-cli

vault-cli is a vault automation tool, used to configure a vault server
with all of the namespaces, endpoints, policies, roles auth endpoins, etc.

vault-cli stores its state in convienent yaml format.  This allows a company to
maintain configuration control over the contents of a vault server.

## Try it out

This example uses namespaces. You will need to download Vault Enterprise

[Download](https://releases.hashicorp.com/vault/1.6.3+ent/)

In first terminal window

```bash
vault server -dev -dev-root-token-id root -dev-listen-address 127.0.0.1:8200
```

In second terminal

```bash
git clone https://github.com/ibm/vault-cli
cd vault-cli
go mod vendor
go build
```

The sample files for these examples are located here: [samples](hack/sample)

```bash
export VAULT_NAME=local
export VAULT_TOKEN=root
export VAULT_NAMESPACE=root
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_LOGIN_NAMESPACE=root

./vault-cli put vaultnamespace -c=ns-test "*"
./vault-cli put vaultauth -c=ns-test "*"
./vault-cli put vaultendpoint -c=ns-test "*"
./vault-cli put vaultpolicy -c=ns-test "*"
./vault-cli put vaultrole -c=ns-test "*"
./vault-cli put jwtrole -c=ns-test "*"
./vault-cli put pkirole -c=ns-test "*"
./vault-cli put sshrole -c=ns-test "*"

vault namespace list -namespace=root
vault namespace list -namespace=parent
vault auth list -namespace=parent
vault policy read -namespace=parent pki-admin
vault read -namespace=parent /auth/jwt/role/operator
vault read -namespace=root /pki/roles/tls
vault read -namespace=root /ssh/roles/operator
vault read -namespace=parent /auth/myauth/role/operator
```

## templates

```bash
# templates
./vault-cli put vaultnamespace -c=tpl-test "*"

vault namespace list -namespace=root

```
