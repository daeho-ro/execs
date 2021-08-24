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

type Execs struct {
    client   ecs.Client
    cluster  string
    task     string
    runtime  string
    region   string
    endpoint string
}

func GetEcsSession(p *Execs) {

    cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(p.region))
    if err != nil {
        log.Fatalf("Unable to load SDK config, %v", err)
    }

    p.client = *ecs.NewFromConfig(cfg)
}

func GetCluster(p *Execs) {

    listClusters, err := p.client.ListClusters(context.TODO(), &ecs.ListClustersInput{})
    if err != nil {
        log.Fatalf("Failed to list ecs clusters, %v", err)
    }

    var clusters []string
    for _, arn := range listClusters.ClusterArns {
        clusterName := strings.Split(arn, "/")[1]
        clusters = append(clusters, clusterName)
    }

    p.cluster = SelectCluster(clusters)
}

func GetTask(p *Execs) {

    listTasks, err := p.client.ListTasks(context.TODO(), &ecs.ListTasksInput{
        Cluster: &p.cluster,
    })
    if err != nil {
        log.Fatalf("Failed to list ecs tasks, %v", err)
    }

    listTaskDetails, err := p.client.DescribeTasks(context.TODO(), &ecs.DescribeTasksInput{
        Tasks: listTasks.TaskArns,
        Cluster: &p.cluster,
    })
    if err != nil {
        log.Fatalf("Failed to describe ecs tasks, %v", err)
    }

    var tasks []string
    for _, task := range listTaskDetails.Tasks {
        taskId := strings.Split(*task.TaskArn, "/")[2]
        taskDefinition := strings.Split(*task.TaskDefinitionArn, "/")[1]
        runtime := *task.Containers[0].RuntimeId
        tasks = append(tasks, fmt.Sprintf("%s | %s | %s", taskId, taskDefinition, runtime))
    }

    var taskInfo = strings.Split(SelectTask(tasks), " | ")
    p.task    = taskInfo[0]
    p.runtime = taskInfo[2]
}

func RunExecuteCommand(p *Execs) {

    var command = "/bin/sh"

    output, err := p.client.ExecuteCommand(context.TODO(), &ecs.ExecuteCommandInput{
        Cluster: &p.cluster,
        Task: &p.task,
        Interactive: true,
        Command: &command,
    })
    if err != nil {
        log.Fatalf("Failed to execute command, %v", err)
    }

    session, err := json.Marshal(output.Session)
    if err != nil {
        log.Fatalf("Json marshal for session is wrong, %v", err)
    }

    var target = fmt.Sprintf("ecs:%s_%s_%s", p.cluster, p.task, p.runtime)
    var ssmTarget = ssm.StartSessionInput{
        Target: &target,
    }
    targetJson, err := json.Marshal(ssmTarget)
    if err != nil {
        log.Fatalf("Json marshal for ssmTarget is wrong, %v", err)
    }

    cmd := exec.Command("session-manager-plugin", string(session), p.region, "StartSession", "", string(targetJson), p.endpoint)
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
