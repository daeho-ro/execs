package main

import (
    "context"
    "fmt"
    "log"
    "strings"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ecs"
    "github.com/manifoldco/promptui"
    // "github.com/AlecAivazis/survey/v2"
)

func main() {

    var cluster = selectCluster()
    var task = selectTask(cluster)
    fmt.Println(task)
}


func getEcsSession() (*ecs.Client) {

    cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-2"))
    if err != nil {
        log.Fatalf("unable to load SDK config, %v", err)
    }

    svc := ecs.NewFromConfig(cfg)

	return svc
}


func selectCluster() (string) {

	svc := getEcsSession()

    listClusters, err := svc.ListClusters(context.TODO(), &ecs.ListClustersInput{})

    if err != nil {
        log.Fatalf("Failed to list ecs clusters, %v", err)
    }

    var clusters []string
    for _, arn := range listClusters.ClusterArns {
        clusters = append(clusters, strings.Split(arn, "/")[1])
    }

    prompt := promptui.Select{
		Label: "Select ECS Cluster",
		Items: clusters,
	}

	_, cluster, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "error"
	}

	return cluster
}


func selectTask(cluster string) (string) {

	svc := getEcsSession()

    listTasks, err := svc.ListTasks(context.TODO(), &ecs.ListTasksInput{
        Cluster: &cluster,
    })

    if err != nil {
        log.Fatalf("Failed to list ecs tasks, %v", err)
    }

    listTaskDetails, err := svc.DescribeTasks(context.TODO(), &ecs.DescribeTasksInput{
        Tasks: listTasks.TaskArns,
        Cluster: &cluster,
    })

    if err != nil {
        log.Fatalf("Failed to describe ecs tasks, %v", err)
    }

    var tasks []string
    for _, task := range listTaskDetails.Tasks {
        tasks = append(tasks, strings.Split(*task.TaskArn, "/")[2] + " - " + strings.Split(*task.TaskDefinitionArn, "/")[1])
    }

    prompt2 := promptui.Select{
		Label: "Select ECS Task",
		Items: tasks,
	}

	_, task, err := prompt2.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "error"
	}

	return task
}
