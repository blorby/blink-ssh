# blink-ssh
The blink-ssh plugin allows the user to perform SSH through Blink. Secure Shell (SSH) is a cryptographic network protocol for operating network services securely over an unsecured network. Typical applications include remote command-line, login, and remote command execution, but any network service can be secured with SSH.

## Connection
In order to use the blink-ssh plugin, you will need to have the following credentials which are required to perform SSH:

- **Username** - The username of the user to connect to on the machine.
- **SSH Private Key** (typically begins with:-----BEGIN RSA PRIVATE KEY----- and ends with:END RSA PRIVATE KEY-----).
- **Passphrase** (optional) - The passphrase is used to encrypt the SSH private key.
- **Host** - The host to connect to, can be an IP address or hostname.