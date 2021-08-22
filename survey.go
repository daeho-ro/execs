package main

import (
    "github.com/AlecAivazis/survey/v2"
)

func SelectCluster(clusters []string) (string) {

    prompt := &survey.Select{
		Message: "Select ECS Cluster",
		Options: clusters,
	}

    var cluster string
	survey.AskOne(prompt, &cluster)
    
	return cluster
}


func SelectTask(tasks []string) (string) {

    prompt := &survey.Select{
		Message: "Select ECS Task",
		Options: tasks,
	}

    var task string
	survey.AskOne(prompt, &task)
    
	return task
}
