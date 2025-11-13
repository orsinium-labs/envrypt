# envrypt

CLI tool for encrypting production secrets.

The secrets are stored in a YAML file with values encrypted with a single password. The file can be decrypted into a `.env` file that can be used with anything, from custom shell scripts to k8s.

Features:

* No external dependencies, no GPG keys, just one plaintext password.
* Allows non-secret values to be stored unencrypted.
* Single binary.
* Cross-platform.
* Works great with `git blame` and `git diff`: you can see exactly who changed what.
* Preserves the original file formatting, variable order, and comments.

## Installation

Can be installed using Go:

```bash
go install github.com/orsinium-labs/envrypt@latest
```

Or you can download a binary release from the [Releases page](https://github.com/orsinium/envrypt/releases).

## Usage

Create `envrypt.yaml` file:

```yaml
# database settings
DB_USER:
    plain: postgres
DB_PASS:
    encrypt: very-secret-very-password
DB_URL:
    plain: "postgres://${DB_USER}:${DB_PASS}@localhost:5432"
```

Encrypt:

```bash
export PASSWORD="keep it secret, keep it safe"
envrypt encrypt --password=$PASSWORD
```

This will encrypt all values marked as `encrypt`. Now the file content will be:

```yaml
# database settings
DB_USER:
    plain: postgres
DB_PASS:
    encrypted: wy4ECQMIeK7BXSTAYm<...>
DB_URL:
    plain: "postgres://${DB_USER}:${DB_PASS}@localhost:5432"
```

Decrypt:

```bash
envrypt decrypt --password=$PASSWORD
```

This will generate `.env` file with decrypted values:

```sh
# database settings
export DB_USER="postgres"
export DB_PASS="very-secret-very-password"
export DB_URL="postgres://${DB_USER}:${DB_PASS}@localhost:5432"
```
