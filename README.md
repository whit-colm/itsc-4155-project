# itsc-4155-project

[SP25 UNCC ITSC-4155:051] Group work monorepo 

# Information

> If you need access to the Google Drive, Figma, or Taiga, or need additional permissions, please reach out via email or Discord.

- [Discord Server](https://discord.com/invite/sQQUmxj8Dp)
- [Google Drive (UNCC read-only)](https://drive.google.com/drive/folders/185QfSHVAMWXiWCKvke5479m5-zHcsPNh?usp=sharing)
- [Figma](https://www.figma.com/files/team/1470848791941601365/all-projects)
- [Taiga](https://tree.taiga.io/project/ailevbar-itsc-4155-spring-2025-team-9)

# Running - Docker

> [!WARNING]
> **USE A CONTAINER!** Production code is not intended to just run directly.

Due to the large array of tooling used in the project, it is intended to be run as a\* docker container. There's an asterisk in the previous sentence because not all of the system can run within a single container; a database is required for persistent storage, which is generally provided as a container as well.

> <details><summary><em>"Why are you doing this to me?"</em> - An explanation of Docker</summary>
> 
> Docker containers are only really meant to run one single program at a time. In this project we actually have 3: the frontend (nginx), the backend (the compiled Go binary), and the database (postgres). For sake of everyone's mental wellness the front and backends can be stapled together with minimal modification, but the database is so grotesquely complex that it can't be included in a monolith Docker container
>
> In a cloud-native environment, like the one this project is designed for, these containers are managed by an even more mind-numbing, nightmare-inducing, megalithic system called a "container orchestrator", the most popular one of these being Kubernetes.
> 
> In Kubernetes, this program would be deployed as a set of pods (where a 'pod' is a group of Docker containers): the front and back-end would be made into their own containers and placed together in a pod, and then the database would be its own pod. The three would communicate over an internal network with a single entry point from the internet (or LAN) which would be the frontend.
>
> Whit (the person writing this) has a Kubernetes cluster at home which if she were not so lazy could set up automatic deployment, but she is so... sorry.
> 
> </details>

## Installing Docker

For _almost_ everybody, you want to install [Docker Desktop](https://docs.docker.com/desktop/). If you are using Windows, make sure you set up [Linux Containers on Windows](https://learn.microsoft.com/en-us/virtualization/windowscontainers/deploy-containers/set-up-linux-containers)

## Testing Monoscript

Once you are sure Docker is installed, you can use the following scripts to build and start the testing environment.

> [!IMPORTANT]
> This has only been fully tested on amd64 Linux. The scripts for Windows and macOS were translated from the Linux script with the help of DeepSeek-R1.
>
> Furthermore, there could be additional complications for users of arm64 (such as Apple Silicon)

In the `/scripts` directory are three scripts:

- `run-test-env.sh` for Linux and other UNIX-like platforms (Bash)
- `RunTestEnv.ps1` for Windows (PowerShell)
- `run-test-env.zsh` for macOS (ZSH)

Note that this script should only be used for testing and demo purposes. Full deployments must be done manually as of now, see below for more information:

## Building

> [!NOTE]
>
> TODO: Write this

## Running

> [!NOTE]
>
> TODO: Write this

## Deploying

> [!NOTE]
>
> TODO: Write this or scrap it.

# Organization

## Frontend

Frontend website tooling is located [in the `/website` subdirectory`](website/).

## Backend

This project's structure is informed by the (unofficial) [golang project layout](https://github.com/golang-standards/project-layout). This section will be updated with further details in future.