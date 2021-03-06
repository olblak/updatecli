source:
  kind: githubRelease
  spec:
    owner: olblak
    repository: updatecli
    token: {{ requiredEnv "GITHUB_TOKEN" }}
    username: olblak
    version: latest
conditions:
  dockerFile:
    name: isDockerfileCorrect
    kind: dockerfile
    spec:
      file: Dockerfile
      Instruction:
        keyword: "FROM"
        matcher: "golang:1.14"
    scm:
      github:
        user: "updatecli"
        email: "updatecli@olblak.com"
        owner: "olblak"
        repository: "updatecli"
        token: {{ requiredEnv "GITHUB_TOKEN" }}
        username: "olblak"
        branch: main
targets:
  dockerFile:
    name: "isDockerfileCorrect"
    kind: dockerfile
    prefix: "olblak/updatecli:"
    spec:
      file: Dockerfile
      Instruction:
        keyword: "FROM"
        matcher: "golang:1.14"
    scm:
      github:
        user: "updatecli"
        owner: "olblak"
        repository: "updatecli"
        token: {{ requiredEnv "GITHUB_TOKEN" }}
        username: "olblak"
        branch: main
        email: "update-bot@olblak.com"
