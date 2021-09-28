package main

import (
	"errors"
	"log"
	"time"
)

func findApplication(applicationName string) (Application, error) {
	repository := initializeRepository(true)

	for _, application := range repository.applications {
		if applicationName == application.name {
			return application, nil
		}
	}

	log.Fatal("No application found with name: %s", applicationName)
	return Application{}, errors.New("No application found")
}

func single(action string, applicationName string) error {
	startTime := time.Now()

	application, err := findApplication(applicationName)

	if err != nil {
		return err
	}

	var commands []string
	switch action {
	case "check":
		commands = application.commands.check
	case "build":
		commands = application.commands.build
	case "analyze":
		commands = application.commands.analyze
	case "release":
		commands = application.commands.release
	case "migrate":
		commands = application.commands.migrate
	case "postrelease":
		commands = application.commands.postrelease
	default:
		log.Fatal("Action is not available")
	}

	for _, jobCommand := range commands {
		output := executeCommand(jobCommand, application.path)
		if IsTrace() {
			log.Printf("%s", output)
		}
	}

	endTime := time.Now()
	timeTaken := endTime.Sub(startTime)
	log.Printf("Total time taken: %s", timeTaken)

	return nil
}
