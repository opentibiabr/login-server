# OpenTibiaBR - Login Server

[![Version](https://img.shields.io/github/v/release/opentibiabr/login-server)](https://github.com/opentibiabr/login-server/releases/latest)
[![Go](https://img.shields.io/github/go-mod/go-version/opentibiabr/login-server)](https://golang.org/doc/go1.16)
![GitHub repo size](https://img.shields.io/github/repo-size/opentibiabr/login-server)

[![Discord Channel](https://img.shields.io/discord/528117503952551936.svg?style=flat-square&logo=discord)](https://discord.gg/3NxYnyV)
[![GitHub pull request](https://img.shields.io/github/issues-pr/opentibiabr/login-server)](https://github.com/opentibiabr/login-server/pulls)
[![GitHub issues](https://img.shields.io/github/issues/opentibiabr/login-server)](https://github.com/opentibiabr/login-server/issues)


## Project

OpenTibiaBR - Login Server is a free open source login server developed in golang to enable cipclient and [otclient](https://github.com/opentibiabr/otclient) to connect and login to [canary server](https://github.com/opentibiabr/canary).

Current version supports only http login, through `/`, `/login`, or `/login.php` routes.

The project is fully covered by tests and supports multi-platform build.
Every release is available with multi-platform applications for download.

## Builds
| Platform       | Build        |
| :------------- | :----------: |
| MacOS          | [![MacOS Build](https://github.com/opentibiabr/login-server/actions/workflows/ci-build-macos.yml/badge.svg?branch=main)](https://github.com/opentibiabr/login-server/actions/workflows/ci-build-macos.yml)   |
| Ubuntu         | [![Ubuntu Build](https://github.com/opentibiabr/login-server/actions/workflows/ci-build-ubuntu.yml/badge.svg?branch=main)](https://github.com/opentibiabr/login-server/actions/workflows/ci-build-ubuntu.yml) |
| Windows        | [![Windows Build](https://github.com/opentibiabr/login-server/actions/workflows/ci-build-windows.yml/badge.svg?branch=main)](https://github.com/opentibiabr/login-server/actions/workflows/ci-build-windows.yml) |

[![Workflow](https://github.com/opentibiabr/login-server/actions/workflows/ci-multiplat-release.yml/badge.svg)](https://github.com/opentibiabr/login-server/actions/workflows/ci-multiplat-release.yml)

### Getting **Started**

To run it, simply download the latest release and define your environment variables.
You can set environment type as `dev` if you want to use a `.env` file (store it in the same folder of the login server).

You can also download our docker image and apply the environment variables to your container.

**Enviroment Variables**

|       NAME          |            HOW TO USE                |
| :------------------ | :----------------------------------  |
|`MYSQL_DBNAME`       | `database default database name`     |
|`MYSQL_HOST`         | `database host`                      |
|`MYSQL_PORT`         | `database port`                      |
|`MYSQL_PASS`         | `database password`                  |
|`MYSQL_USER`         | `database username`                  |
|`ENV_LOG_LEVEL`      | `logrus log level for verbose` [ref](https://pkg.go.dev/github.com/sirupsen/logrus#Level)   |
|`ENV_LOG_FILE`       | `optional log file path, defaults to logs/login-server.txt` |
|`LOGIN_IP`           | `login ip address`                   |
|`LOGIN_HTTP_PORT`    | `login http port`                    |
|`LOGIN_GRPC_PORT`    | `login grpc port`                    |
|`LOGIN_TRUSTED_PROXIES` | Comma-separated IPs/CIDRs of the reverse proxies allowed to supply `X-Forwarded-For` / `X-Real-IP` (and `X-Forwarded-Proto`). Empty by default: forwarded headers are ignored and the TCP peer address is used. Behind a proxy on the same host set `127.0.0.1`. The proxy must append the client address to `X-Forwarded-For` (nginx: `$proxy_add_x_forwarded_for`). An invalid entry discards the whole list. |
|`LOGIN_FREE_PREMIUM` | `true` (or `1`, `t`, `True`: any value Go's `strconv.ParseBool` reads as true) shows an account without paid time as a free premium account in the client character list; use it when the game server runs with freePremium. Unset, empty or any other value shows it as a free account. An account with paid time (`accounts.lastday` in the future) shows its whole days left, rounded down and never less than one while any paid time is left|
|`LOGIN_TRUSTED_PROXIES`|`comma-separated IP/CIDR allowlist of reverse proxies permitted to supply forwarding headers; empty by default`|
|`RATE_LIMITER_BURST` | `rate limiter same request burst`    |
|`RATE_LIMITER_RATE`  | `rate limit request per sec per user`|
|`AUTHENTICATOR_ENCRYPTION_KEY`|`base64-encoded 32-byte AES key shared with the account website; required at login-server startup`|
|`SERVER_IP`          | `game server IP address`             |
|`SERVER_LOCATION`    | `game server location`               |
|`SERVER_PVP_TYPE`    | world PvP type in the world list (`pvptype`): a `worldType` name (`retro-pvp`, `pvp`, `no-pvp`, `pvp-enforced`, `retro-pvp-enforced`, `expert-pvp`) or a client code 0 to 4. Unset: `worldType` from the `config.lua` of `SERVER_PATH`; without either, or unknown, open PvP |
|`SERVER_NAME`        | `game server name; the official client sends this exact value plus newline before the first world-login packet` |
|`SERVER_PORT`        | `game server game port`              |
|`VOCATIONS`          | `game vocation list csv (a,b,c)`     |

HTTP login aliases `/`, `/login`, and `/login.php` separate authentication traffic
from the public `cacheinfo`, `news`, `newsviewed`, `eventschedule`, and
`boostedcreature` requests. Each IP has one authentication bucket and one shared
public bucket, both using `RATE_LIMITER_RATE` and `RATE_LIMITER_BURST`. Public
bootstrap requests therefore cannot consume the authentication burst of clients
sharing an IP. Malformed payloads, unsupported request types, and other routes
remain in the authentication bucket. Idle buckets expire after three minutes.

Login JSON is decoded once and shared between admission and dispatch. Its body is
limited to 1 MiB; invalid or oversized JSON retains the HTTP 400 JSON error format.
When neither request class can admit a request, the server returns HTTP 429 before
reading the JSON body. The capacity check does not consume quota; the selected
class's final admission remains authoritative under concurrent requests. The HTTP
server bounds header reads to five seconds and complete request reads to 30
seconds, including slow bodies that begin while capacity is available.

> **Behind a reverse proxy, set `LOGIN_TRUSTED_PROXIES`.** Forwarded client IP headers are no longer trusted by default. Without it, every request that arrives through a proxy is attributed to the proxy address, so all clients share a single rate limit bucket (`RATE_LIMITER_BURST` / `RATE_LIMITER_RATE`) and a burst of logins, such as every client reconnecting after a restart, is answered with 429. Trusting every peer, the old default, let a client choose its own address: gin returns the leftmost `X-Forwarded-For` entry, which the client controls, so `X-Forwarded-For: 9.9.9.9` arrives as `9.9.9.9, <real address>` and is rate limited as `9.9.9.9`. With the proxy listed, gin reads the header from the right and uses the address the proxy appended.

**Tests**  
`go test ./...`

**Build**  
`RUN go build -o TARGET_NAME ./src/`

## Account passwords

`accounts.password` is read by its own format, so accounts in different formats work side by side:

| Stored value | Meaning |
|---|---|
| `$argon2id$v=19$m=65536,t=3,p=1$<salt>$<hash>` | Argon2id of the password, as written by PHP `password_hash` or libargon2. The parameters are read from the value (limits: `m` up to 262144 KiB, `t` and `p` up to 16) |
| `sha1:$argon2id$...` | Argon2id of the lowercase hexadecimal SHA-1 of the password: a former SHA-1 value wrapped without knowing the password |
| 40 hexadecimal characters | Legacy SHA-1 of the password |

|NAME|HINT|
|---|---|
|`LOGIN_PASSWORD_REHASH`| `true` replaces a SHA-1, wrapped or weaker Argon2id value (less memory, or the same memory and fewer passes) with `m=65536,t=3,p=1` after a successful password check. Default `false`. A failed update is logged and does not fail the login. Nothing is written, and nothing is hashed, while `accounts.password` holds fewer than 128 characters (widen it first, see below) |
|`LOGIN_PASSWORD_CHANGE_MESSAGE`| Message shown (error `PASSWORD_CHANGE_REQUIRED`, code 8) when the password is correct but `accounts.password_change_required` is `1`. Default: `You must change your account password on the website before you can play.` The check exists only when that column is present (detected once at startup); without it logins are unchanged. A wrong password still answers the normal invalid-credentials error, and no session, trusted-device token or password upgrade is issued for a flagged account |
|`LOGIN_PASSWORD_HASH_CONCURRENCY`| Argon2id computations that may run at once (1 to 64, default 2). A login that cannot get a slot within 10 seconds answers `PASSWORD_VERIFICATION_BUSY` (LS-3005), which is not a credentials error |

Memory: each computation uses the `m` of the value it checks (64 MiB for the values written here). It is released before the slot is handed over, but the runtime does not return it at once: plan for that amount per slot and one more. On a service limited to 256 MiB, set the concurrency to 1, and keep stored values at `m=65536`. One request computes Argon2id for at most 8 accounts sharing the same email or name. One computation takes on the order of 100 ms of one CPU core.

Enable `LOGIN_PASSWORD_REHASH` only when everything else that reads `accounts.password` accepts Argon2id (the game server, when it runs with `authType = "password"` or the old protocol, and the account website), widen the column if it is narrower than 128 characters (for example `ALTER TABLE accounts MODIFY password VARCHAR(255) NOT NULL;`, keeping the column's own nullability and default), and give the login-server database user `UPDATE` on that column:

```sql
GRANT UPDATE (`password`) ON `canary`.`accounts` TO 'login'@'localhost';
```

After values have been rewritten, do not roll back to a version that only reads SHA-1: those accounts could not log in.

## Two-factor authentication

The login endpoint accepts an optional six-digit `token` field. Accounts with a row in `account_authenticators` must provide a valid RFC 6238 TOTP code. Account types 4 and higher, and accounts that own a character that is not deleted with `players.group_id` 4 or higher, must enroll before they can log in; enrollment remains optional for other accounts. For an account without an authenticator, a failure to read its characters' groups refuses the login.

The account website owns enrollment and stores the secret only after the player confirms the first code. The website and login server must use the same `AUTHENTICATOR_ENCRYPTION_KEY`. Generate it once with `openssl rand -base64 32`, store it outside the database, and do not rotate it without re-encrypting enrolled secrets.

The game server must run with `authType = "session"`. The login API refuses an explicit password-authentication configuration because returning `email\npassword` would let direct game-server logins bypass the authenticator.

The required table is:

```sql
CREATE TABLE `account_authenticators` (
  `account_id` int unsigned NOT NULL,
  `secret_encrypted` varchar(512) NOT NULL,
  `last_used_step` bigint unsigned DEFAULT NULL,
  `failed_attempts` smallint unsigned NOT NULL DEFAULT 0,
  `blocked_until` bigint unsigned NOT NULL DEFAULT 0,
  `enabled_at` bigint unsigned NOT NULL,
  PRIMARY KEY (`account_id`),
  CONSTRAINT `account_authenticators_account_fk`
    FOREIGN KEY (`account_id`) REFERENCES `accounts` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

Secrets use the envelope `v1:<base64(nonce || ciphertext || tag)>`, with AES-256-GCM, a 12-byte nonce, a 16-byte tag, and associated data `otbr-login-authenticator:v1:<account_id>`. Tokens use SHA-1, six digits, a 30-second period, and a one-step clock window. Accepted time steps are recorded atomically so the same code cannot be reused. Invalid attempts are limited per account in addition to the HTTP per-IP limiter.

HTTPS clients may send `trustdevice: true` with a successful TOTP challenge. The response then includes an opaque `trusteddevicetoken` that can replace TOTP for that account for 30 days; the account password is still required on every login. The credential is rotated after every successful use, only its SHA-256 hash is stored, and at most ten active devices are kept per account. The API ignores enrollment and trusted-device credentials unless the request uses TLS directly or carries `X-Forwarded-Proto: https` from a peer listed in `LOGIN_TRUSTED_PROXIES`. Clients must protect the token with operating-system credential protection and must never send it over plain HTTP.

When TLS terminates at an upstream proxy such as Cloudflare, production deployments must also use authenticated TLS from that proxy to the login server and validate the origin certificate; use mutual TLS when the proxy supports it. `X-Forwarded-Proto` and an IP allowlist authenticate neither the transport nor its contents and do not protect the account password or trusted-device token on that hop. Plain HTTP is acceptable only on same-host loopback or a private network that is isolated from untrusted workloads and packet observers, with that exception recorded in the deployment threat model.

Set `LOGIN_TRUSTED_PROXIES` to the immediate reverse proxy's IP addresses or CIDR ranges. Configure every listed proxy to remove any client-supplied `X-Forwarded-Proto` and write exactly one authenticated upstream value; the login server rejects repeated or comma-separated values. Do not replace that value with Nginx's local `$scheme`; preserve it only for authenticated and allowlisted proxy addresses, and fall back to `$scheme` for direct origin traffic. When Nginx's RealIP module rewrites `$remote_addr`, classify the proxy peer with `$realip_remote_addr`; otherwise the allowlist is evaluated against the visitor address and every forwarded scheme is rejected. Never trust a client-supplied forwarding header from an unrestricted source. Cloudflare publishes both its [request-header contract](https://developers.cloudflare.com/fundamentals/reference/http-headers/) and its current [origin-facing IP ranges](https://developers.cloudflare.com/fundamentals/concepts/cloudflare-ip-addresses/).

Changing or recovering the account password, enabling, replacing, or disabling the authenticator, and explicit device removal must delete the affected rows from `account_trusted_devices`. The account website exposes individual and all-device revocation. Expiry is absolute and is not extended by use.

## Livestream spy targets

GOD accounts (type 6) that passed second-factor verification get extra `Spy: <name>` entries after their real characters, one per saved target, including offline ones. The list is read from `account_livestream_spy_targets` (at most 64 targets, ordered by name, non-deleted players only); ordinary accounts never query it. If the table is missing the server logs a warning and omits the entries, so normal login keeps working.

```sql
CREATE TABLE `account_livestream_spy_targets` (
  `account_id` int unsigned NOT NULL,
  `target_id` int NOT NULL,
  PRIMARY KEY (`account_id`, `target_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

`target_id` is `players.id`. A target whose `Spy: <name>` would collide with a real character of the same GOD account is omitted.

## Docker
`docker pull opentibiabr/login-server:latest`<br><br>
[![Automation](https://img.shields.io/docker/cloud/automated/opentibiabr/login-server)](https://hub.docker.com/r/opentibiabr/login-server)
[![Image Size](https://img.shields.io/docker/image-size/opentibiabr/login-server)](https://hub.docker.com/r/opentibiabr/login-server/tags?page=1&ordering=last_updated)
![Pulls](https://img.shields.io/docker/pulls/opentibiabr/login-server)
[![Build](https://img.shields.io/docker/cloud/build/opentibiabr/login-server)](https://hub.docker.com/r/opentibiabr/login-server/builds)

## Benchmark
There are a few known login versions available. The most common ones are a python login and login.php, from different websites versions.
We've performed a benchmark (code available in benchmark_test.go) where 1k valid requests were performed to the server, locally.
As you can see in the results below, we got up to 10x faster without decreasing the server availability (99.5% is still pretty good in any global standard).

![image](https://user-images.githubusercontent.com/34237492/118380499-7da2f500-b5e2-11eb-9025-eda180d501df.png)

Also, we performed a benchmark in google cloud, using cloud run and cloud sql database, both with lower possible specifications.
As you can see, we kept an average of 700 requests/s and a good availability, even if with the latency between my computed and cloud services being accounted in this graph.

![image](https://user-images.githubusercontent.com/34237492/118379403-64964600-b5da-11eb-9e11-25c92024986d.png)

Another great aspect is that, comparing with the python login, our docker image is almost 10x smaller (15Mb). 

## gRPC
From version 2.0.0 on, we start using gRPC protocol. 
The HTTP runs on top of the gRPC layer, using a reversed proxy. That lead to a small gain in the availability without any performance loss.

In the gRPC server we got a 10x performance boost, compared to the HTTP benchmarks:

![image](https://user-images.githubusercontent.com/34237492/118568814-e45a1700-b778-11eb-8b79-ddc26dde487c.png)

## Issues

We use the [issue tracker on GitHub](https://github.com/opentibiabr/login-server/issues). Everyone who is watching the repository gets notified by e-mail when there is an activity, so be mindful about comments that add no value (e.g. "+1"). 

We are willing to improve the login server with more features, so feel free to create issues with features requests and ideas, only bug fixes.

If you'd need an issue/feature to be prioritized, you should either do it yourself and submit a pull request, or place a bounty.

## Pull requests

Before [creating a pull request](https://github.com/opentibiabr/login-server/pulls) please keep in mind:

* Set one single scope in your pull request. Focus help us review and things to ship faster. Too many things on the same Pull Request make it harder to review, harder to test and hard to move on.
* Add tests. Pull Requests without tests **won't** be approved.
* Your code must follow go [standard golang format patterns](https://golang.org/doc/effective_go#formatting).
* There are people that doesn't play the game on the official server, so explain your changes to help understand what are you changing and why.
* Avoid opening a Pull Request to just update minor typo or comments. Try attaching those to other PRs with meaningful content.

## Special Thanks

* our partners
* our crew (majesty, gpedro, eduardo dantas, foot, lucas)

## **Sponsors**

If you want to sponsor here, join on discord and send a message for one of our administrators.

## Partners

[![Supported by OTServ Brasil](https://raw.githubusercontent.com/otbr/otserv-brasil/main/otbr.png)](https://forums.otserv.com.br)
