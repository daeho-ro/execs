package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
    "os"
    "os/exec"
    // "os/signal"
	"strings"
    // "syscall"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type session struct {

}

func GetEcsSession(region string) (*ecs.Client) {

    cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
    if err != nil {
        log.Fatalf("Unable to load SDK config, %v", err)
    }

    svc := ecs.NewFromConfig(cfg)

	return svc
}

func GetClusters(svc *ecs.Client) ([]string) {

    listClusters, err := svc.ListClusters(context.TODO(), &ecs.ListClustersInput{})
    if err != nil {
        log.Fatalf("Failed to list ecs clusters, %v", err)
    }

    var clusters []string
    for _, arn := range listClusters.ClusterArns {
        clusters = append(clusters, strings.Split(arn, "/")[1])
    }

    return clusters
}


func GetTasks(svc *ecs.Client, cluster string) ([]string) {

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
        tasks = append(tasks, fmt.Sprintf("%s | %s | %s", strings.Split(*task.TaskArn, "/")[2], strings.Split(*task.TaskDefinitionArn, "/")[1], *task.Containers[0].RuntimeId))
    }

    return tasks
}

func RunExecuteCommand(svc *ecs.Client, cluster string, task string, runtime string) {

    var command = "/bin/sh"

    output, err := svc.ExecuteCommand(context.TODO(), &ecs.ExecuteCommandInput{
        Cluster: &cluster,
        Task: &task,
        Interactive: true,
        Command: &command,
    })
    if err != nil {
        log.Fatalf("Failed to describe ecs tasks, %v", err)
    }

    session, err := json.Marshal(output.Session)
    if err != nil {
        log.Fatalf("something wrong, %v", err)
    }


    var target = fmt.Sprintf("ecs:%s_%s_%s", cluster, task, runtime)
    var ssmTarget = ssm.StartSessionInput{
        Target: &target,
    }
    targetJson, err := json.Marshal(ssmTarget)
    if err != nil {
        log.Fatalf("something wrong, %v", err)
    }

    // var args = fmt.Sprintf(

    // cmd := exec.Command("aws", "ecs", "execute-command", "--cluster", cluster, "--task", task, "--command", command, "--interactive")
    // cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout
	// cmd.Stdin = os.Stdin

    cmd := exec.Command("session-manager-plugin", string(session), "ap-northeast-2", "StartSession", "", string(targetJson), "https://ssm.ap-northeast-2.amazonaws.com")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

	// sigs := make(chan os.Signal, 1)
	// signal.Notify(sigs, os.Interrupt, syscall.SIGINT)
	// go func() {
	// 	for {
	// 		select {
	// 		case <-sigs:
	// 		}
	// 	}
	// }()
	// defer close(sigs)

    if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
}
