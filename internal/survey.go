package ecs

import (
	"log"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

func selectRegion(regions []string, baseRegion string) string {

	prompt := &survey.Select{
		Message: "Select Region",
		Options: regions,
		Default: baseRegion,
	}

	var region string
	err := survey.AskOne(prompt, &region)
	if err == terminal.InterruptErr {
		log.Fatal("interrupted")
		os.Exit(0)
	} else if err != nil {
		panic(err)
	}

	return region
}

func selectCluster(clusters []string) string {

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

func selectTask(tasks []string) string {

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

func selectContainer(containers []string) string {

	prompt := &survey.Select{
		Message: "Select Container",
		Options: containers,
	}

	var container string
	err := survey.AskOne(prompt, &container)
	if err == terminal.InterruptErr {
		log.Fatal("interrupted")
		os.Exit(0)
	} else if err != nil {
		panic(err)
	}

	return container
}
