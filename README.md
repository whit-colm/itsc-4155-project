# itsc-4155-project

[SP25 UNCC ITSC-4155:051] Group work monorepo 

# Information

> If you need access to the Google Drive, Figma, or Taiga, or need additional permissions, please reach out via email or Discord.

- [Discord Server](https://discord.com/invite/sQQUmxj8Dp)
- [Google Drive (UNCC read-only)](https://drive.google.com/drive/folders/185QfSHVAMWXiWCKvke5479m5-zHcsPNh?usp=sharing)
- [Figma](https://www.figma.com/files/team/1470848791941601365/all-projects)
- [Taiga](https://tree.taiga.io/project/ailevbar-itsc-4155-spring-2025-team-9)

# Running - Docker

> :warning: **USE A CONTAINER!** Production code is not intended to just run directly.

Due to the large array of tooling used in the project,

## Installing Docker

For _almost_ everybody, you want to install [Docker Desktop](https://docs.docker.com/desktop/). If you are using Windows, make sure [you set up Linux Containers on Windows](https://learn.microsoft.com/en-us/virtualization/windowscontainers/deploy-containers/set-up-linux-containers)

## Running

<!-- TODO: These will probably become "dev" and "prod"; but not right now. -->
There are two modes primary targets to meet, `monolith` and `webnative`.

# Organization

## Frontend

Frontend website tooling is located [in the `/website` subdirectory`](website/).

## Backend

This project's structure is informed by the (unofficial) [golang project layout](https://github.com/golang-standards/project-layout). This section will be updated with further details in future.