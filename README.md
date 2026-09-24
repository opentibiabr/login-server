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
|`LOGIN_TRUSTED_PROXIES`|`comma-separated IP/CIDR allowlist of reverse proxies permitted to supply forwarding headers; empty by default`|
|`RATE_LIMITER_BURST` | `rate limiter same request burst`    |
|`RATE_LIMITER_RATE`  | `rate limit request per sec per user`|
|`AUTHENTICATOR_ENCRYPTION_KEY`|`base64-encoded 32-byte AES key shared with the account website; required at login-server startup`|
|`SERVER_IP`          | `game server IP address`             |
|`SERVER_LOCATION`    | `game server location`               |
|`SERVER_NAME`        | `game server name; the official client sends this exact value plus newline before the first world-login packet` |
|`SERVER_PORT`        | `game server game port`              |
|`VOCATIONS`          | `game vocation list csv (a,b,c)`     |

**Tests**  
`go test ./...`

**Build**  
`RUN go build -o TARGET_NAME ./src/`

## Two-factor authentication

The login endpoint accepts an optional six-digit `token` field. Accounts with a row in `account_authenticators` must provide a valid RFC 6238 TOTP code. Account types 4 and higher must enroll before they can log in; enrollment remains optional for other accounts.

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

Set `LOGIN_TRUSTED_PROXIES` to the immediate reverse proxy's IP addresses or CIDR ranges. Do not replace the proxy's `X-Forwarded-Proto` with Nginx's local `$scheme`; preserve the upstream value only for authenticated and allowlisted proxy addresses, and fall back to `$scheme` for direct origin traffic. When Nginx's RealIP module rewrites `$remote_addr`, classify the proxy peer with `$realip_remote_addr`; otherwise the allowlist is evaluated against the visitor address and every forwarded scheme is rejected. Never trust a client-supplied forwarding header from an unrestricted source. Cloudflare publishes both its [request-header contract](https://developers.cloudflare.com/fundamentals/reference/http-headers/) and its current [origin-facing IP ranges](https://developers.cloudflare.com/fundamentals/concepts/cloudflare-ip-addresses/).

Changing or recovering the account password, enabling, replacing, or disabling the authenticator, and explicit device removal must delete the affected rows from `account_trusted_devices`. The account website exposes individual and all-device revocation. Expiry is absolute and is not extended by use.

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
