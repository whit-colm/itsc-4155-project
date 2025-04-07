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

## Testing Composer

Once you are sure Docker is installed, actual running is done with Docker Compose

> [!WARNING]
> This will not work out of the box on Windows!
>
> Before running this for the first time, see the below section.

In the root of the directory (i.e. the folder with the `docker-compose.aml` file):

```
docker compose up
```

You can add certain flags in addition, the most noteworthy of which are as follows:

- `-d` - run in detached (background) mode. 
    - You can stop the containers with `docker stop <container name>`, if you do not know the names of the containers use `docker ps`
- `--no-deps --build` build the images used by the containers each time before running; this is useful for rapid testing of the docker environment

There are cleanup scripts in `/scripts/` which will prune the containers, images, and volume (used by Postgres for some reason).

## Windows

If you are using Windows, you will need to first change certain git config values or else the app will not start. Please follow these steps prior

1. Delete your local repository
    - Commit and push your changes first, of course. If you are working on something rather special you can merge into a new branch, but you need to delete your copy of the repo.
    - You don't actually need to, but following step 2 almost every file will show changes and you should not commit them.
2. In a PowerShell session, run the following command:
    - ```
      git config --global core.autocrlf false
      ```
3. Re-clone the repo.

The underlying issue is that Window's weird default line terminators break certain shell scripts copied into the containers. This isn't really an issue anymore because software development on Windows is just Linux under the hood anyway, but git has it enabled anyway for compatibility.

# Organization

## Frontend

Frontend website tooling is located [in the `/website` subdirectory`](website/).

## Backend

- This project's structure is informed by the (unofficial) [golang project layout](https://github.com/golang-standards/project-layout). This section will be updated with further details in future.
- Keep [naming conventions](https://google.github.io/styleguide/go/decisions.html) in mind for clarity.

### Testing

> [!IMPORTANT]
> Tests **must** pass before a merge is accepted! Tests are run via:
>
> ```bash
> go test -v ./...
> ```

- Go tests should be written in a separate file in the same directory as the source code being tested, but with `_test.go` appended to the name of the source file.
    - For instance, if you are testing `/pkg/sourcecode.go`, the test file should be `/pkg/sourcecode_test.go`.
- You will need to import `"testing"`, potentially along with the package you are testing:
    - If you are **whitebox** testing: the test file's `package` should match the name of the source code you are testing. 
        - e.g. if `/pkg/sourcecode.go` uses `package mycode`, then `/pkg/sourcecode_test.go` should *also* use `package mycode`
    - If you are **blackbox** testing: the test file's `package` should use the name of the source code you are testing, *with `_test` appended*. 
        - e.g. if `/pkg/sourcecode.go` uses `package mycode`, then `/pkg/sourcecode_test.go` should use `package mycode_test`
        - You will also need to import the package you are testing in this case in addition to `"testing"`
- Test functions should be named after the original function being tested, but with the `Test` prefix (and optional descriptive suffix for multiple tests); still in `UpperCamelCase`. All test accept a single parameter of `*testing.T`.
    - If you are testing `AddNumbers(a, b int) int`, perform all your tests in the body of `TestAddNumbers(t *testing.T)`; if you need multiple tests, you can have `TestAddNumbersPositive(t *testing.T)`, `TestAddNumbersNegative(t *testing.T)`, etc
    - You do not need assert states; tests pass automatically when the function ends without an `Errorf` method on the argument `*testing.T`.

