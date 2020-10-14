# reversetls

Automatically renews Let's Encrypt certificates and acts as a reverse proxy to the provided hostnames

## Usage:
Pass it a comma separated list of domains. You can pass aliases for a domain as a space separated list between commas (first domain name is used as the target for the reverse proxy)

Use as binary:

`reversetls [options] domain_0 alias_0 ... alias_i, ... domain_i alias_0 ... alias_i`

Use with docker image:
 - `docker pull registry.gitlab.com/miscing/reversetls`
 - `docker run registry.gitlab.com/miscing/reversetls [options] domain alias, ...`

## Notes:
 - Both aliases and domains are checked to be valid url notation.
 - Gets certificates for both domains and aliases.
