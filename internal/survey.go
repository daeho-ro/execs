package ecs

import (
	"log"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

func SelectCluster(clusters []string) string {

	prompt := &survey.Select{
		Message: "Select ECS Cluster",
		Options: clusters,
	}

	var cluster string
	err := survey.AskOne(prompt, &cluster)
	if err == terminal.InterruptErr {
		log.Fatal("interrupted")
		os.Exit(0)
	} else if err != nil {
		panic(err)
	}

	return cluster
}

func SelectTask(tasks []string) string {

	prompt := &survey.Select{
		Message: "Select ECS Task",
		Options: tasks,
	}

	var task string
	err := survey.AskOne(prompt, &task)
	if err == terminal.InterruptErr {
		log.Fatal("interrupted")
		os.Exit(0)
	} else if err != nil {
		panic(err)
	}

	return task
}