package main

import (
  "path"
  "os"
  "log"
  "strings"
  "regexp"

  "github.com/urfave/cli"
)

func createHelmRepositoryIndex(directory string) string {
  return executeCommand("helm repo index .", directory)
}

func updateHelmIndex(cliContext *cli.Context, repository Repository) error {
  generatedPath := path.Join(repository.path, ".generated")
  if ! directoryExists(generatedPath) {
    os.Mkdir(generatedPath, 0777)
  }
  repositoryName := "helm-repo"
  helmRepositoryPath := path.Join(repository.path, ".generated/", repositoryName)
  if ! directoryExists(helmRepositoryPath) {
    repositoryURL := repository.helm.repositoryGitURL
    log.Print(gitCloneRepository(repositoryURL, repositoryName, generatedPath))
  }

  commitMessage := "re-index"
  log.Print(createHelmRepositoryIndex(helmRepositoryPath))
  log.Print(gitCommit(commitMessage, helmRepositoryPath))

  // TODO: need a retry here.. probably for the clone aswell
  err := gitPush(helmRepositoryPath)

  if err != nil {
    return err
  }

  return nil
}

func getLatestCommit(repository Repository) string {
  command := "git ls-remote " + repository.helm.repositoryGitURL + " refs/heads/master"
  commandResult := executeCommand(command, repository.path)

  expression, _ := regexp.Compile("([a-f0-9]{40})")
  commitId := expression.FindString(commandResult)

  return commitId
}

func setupHelmRepository(cliContext *cli.Context, repository Repository) error {
  commitId := getLatestCommit(repository)
  repositoryCommitUrl := strings.Replace(repository.helm.repository, "master", commitId, 1)
  addRepoCommand := "helm repo add --username " + repository.helm.username
  addRepoCommand += " --password " + repository.helm.password
  addRepoCommand += " knega-repo " + repositoryCommitUrl
  executeCommand(addRepoCommand, repository.path)

  executeCommand("helm repo update", repository.path)

  return nil
}

func ociHelmPackageExists(application *Application) bool {
  packageVersion := "1.0.0-" + application.inputsHash
  imagePath := application.helm.repositoryOCI + ":" + packageVersion
  return ociImageExists(imagePath, application.helm.username, application.helm.password)
}

func helmPackageExists(application *Application) bool {
  searchCommand := "helm search repo --version 1.0.0-" + application.inputsHash
  searchCommand += " knega-repo/" + application.name
  result := executeCommand(searchCommand, application.path)

  if strings.Contains(result, "No results found") {
    return false
  }

  if strings.Contains(result, "APP VERSION") {
    return true
  }

  log.Printf("Something went wrong while checking for helm chart search command: %s returned the following results: %s", searchCommand, result)

  return false
}
