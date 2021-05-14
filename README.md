# OpenTibiaBR - Login Server

[![Version](https://img.shields.io/github/v/release/opentibiabr/login-server)](https://github.com/opentibiabr/login-server/releases/latest)
![Go](https://img.shields.io/github/go-mod/go-version/opentibiabr/login-server)
![GitHub repo size](https://img.shields.io/github/repo-size/opentibiabr/login-server)

[![Discord Channel](https://img.shields.io/discord/528117503952551936.svg?style=flat-square&logo=discord)](https://discord.gg/3NxYnyV)
[![GitHub pull request](https://img.shields.io/github/issues-pr/opentibiabr/login-server)](https://github.com/opentibiabr/login-server/pulls)
[![GitHub issues](https://img.shields.io/github/issues/opentibiabr/login-server)](https://github.com/opentibiabr/login-server/issues)

## Project

OpenTibiaBR - Login Server is a free open source login server developed in golang to enable cipclient and [otclient](https://github.com/opentibiabr/otclient) to connect and login to [canary server](https://github.com/opentibiabr/canary).

Current version supports only http login, through `/login` or `/login.php` routes.

The project is fully covered by tests and supports multi-platform build.
Every release is available with multi-platform applications for download.

## Builds
| Platform       | Build        |
| :------------- | :----------: |
| MacOS          | ![MacOS Build](https://github.com/opentibiabr/login-server/actions/workflows/ci-build-macos.yml/badge.svg?branch=main)   |
| Ubuntu         | ![Ubuntu Build](https://github.com/opentibiabr/login-server/actions/workflows/ci-build-ubuntu.yml/badge.svg?branch=main) |
| Windows        | ![Windows Build](https://github.com/opentibiabr/login-server/actions/workflows/ci-build-windows.yml/badge.svg?branch=main) |

![Workflow](https://github.com/opentibiabr/login-server/actions/workflows/ci-multiplat-release.yml/badge.svg)

### Getting **Started**

To run it, simply download the latest release and define your environment variables.
You can set environment type as `dev` if you want to use a `.env` file (store it in the same folder of the login server).

You can also download our docker image and apply the environment variables to your container.

**Enviroment Variables**

|       NAME       |            HOW TO USE             |
| :-------------- | :--------------------------------- |
|`DB_DATABASE`     | `database default database name`  |
|`DB_HOSTNAME`     | `database host`                   |
|`DB_PORT`         | `database port`                   |
|`DB_PASSWORD`     | `database password`               |
|`DB_USERNAME`     | `database username`               |
|`ENV_RUN_SILENT`  | `true/false disables logs`        |
|`ENV_TYPE`        | `test/dev/prod uses .env if 'dev'`|
|`LOGIN_PORT`      | `login http port`                 |
|`SERVER_IP`       | `game server IP address`          |
|`SERVER_LOCATION` | `game server location`            |
|`SERVER_NAME`     | `game server name`                |
|`SERVER_PORT`     | `game server game port`           |

**Tests**  
`go test ./tests -v`

**Build**  
`RUN go build -o TARGET_NAME ./src/`

### Issues

We use the [issue tracker on GitHub](https://github.com/opentibiabr/login-server/issues). Everyone who is watching the repository gets notified by e-mail when there is an activity, so be mindful about comments that add no value (e.g. "+1"). 

We are willing to improve the login server with more features, so feel free to create issues with features requests and ideas, only bug fixes.

If you'd need an issue/feature to be prioritized, you should either do it yourself and submit a pull request, or place a bounty.

### Pull requests

Before [creating a pull request](https://github.com/opentibiabr/login-server/pulls) please keep in mind:

* Set one single scope in your pull request. Focus help us review and things to ship faster. Too many things on the same Pull Request make it harder to review, harder to test and hard to move on.
* Add tests. Pull Requests without tests **won't** be approved.
* Your code must follow go [standard golang format patterns](https://golang.org/doc/effective_go#formatting).
* There are people that doesn't play the game on the official server, so explain your changes to help understand what are you changing and why.
* Avoid opening a Pull Request to just update minor typo or comments. Try attaching those to other PRs with meaningful content.

### Special Thanks

* our partners
* our crew (majesty, gpedro, eduardo dantas, foot, lucas)

### **Sponsors**

If you want to sponsor here, join on discord and send a message for one of our administrators.

### Partners

[![Supported by OTServ Brasil](https://raw.githubusercontent.com/otbr/otserv-brasil/main/otbr.png)](https://forums.otserv.com.br)
