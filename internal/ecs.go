package ecs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
    "os"
    "os/exec"
    "os/signal"
	"strings"
    "syscall"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type Execs struct {
    client   ecs.Client
    step     chan string
    err      chan error
    quit     chan bool
    cluster  string
    task     string
    runtime  string
    region   string
    endpoint string
    command  string
}

func Start() {

    p := &Execs{
        step: make(chan string, 1),
        err: make(chan error, 1),
        quit: make(chan bool, 1),
        region : "ap-northeast-2",
    }

    p.loop()
}

func (p *Execs) loop() {

	go func() {
		for {
			select {
			case step := <- p.step:
				switch step {
                case "getEcsSession":
                    p.getEcsSession()
				case "getCluster":
					p.getCluster()
				case "getTask":
					p.getTask()
				case "runExecuteCommand":
					p.runExecuteCommand()
				default:
					return
				}
			case err := <- p.err:
				log.Fatal(err)
				os.Exit(1)
			}
		}
	}()

    p.step <- "getEcsSession"

    <- p.quit
}

func (p *Execs) getEcsSession() {

    cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(p.region))
    if err != nil {
        p.err <- err
        log.Fatalf("Unable to load SDK config, %v", err)
    }

    p.client = *ecs.NewFromConfig(cfg)
    p.step <- "getCluster"
}

func (p *Execs) getCluster() {

    listClusters, err := p.client.ListClusters(context.TODO(), &ecs.ListClustersInput{})
    if err != nil {
        p.err <- err
        log.Fatalf("Failed to list ecs clusters, %v", err)
    }
    if len(listClusters.ClusterArns) == 0 {
        log.Printf("There is no ECS clusters in the %s region", p.region)
        p.quit <- true
    }

    var clusters []string
    for _, arn := range listClusters.ClusterArns {
        clusterName := strings.Split(arn, "/")[1]
        clusters = append(clusters, clusterName)
    }

    p.cluster = SelectCluster(clusters)
    p.step <- "getTask"
}

func (p *Execs) getTask() {

    listTasks, err := p.client.ListTasks(context.TODO(), &ecs.ListTasksInput{
        Cluster: &p.cluster,
    })
    if err != nil {
        p.err <- err
        log.Fatalf("Failed to list ecs tasks, %v", err)
    }
    if len(listTasks.TaskArns) == 0 {
        log.Printf("There is no ECS Task in the cluster %s", p.cluster)
        p.step <- "getCluster"
    }

    listTaskDetails, err := p.client.DescribeTasks(context.TODO(), &ecs.DescribeTasksInput{
        Tasks: listTasks.TaskArns,
        Cluster: &p.cluster,
    })
    if err != nil {
        p.err <- err
        log.Fatalf("Failed to describe ecs tasks, %v", err)
    }

    var tasks []string
    for _, task := range listTaskDetails.Tasks {
        if *task.LastStatus == "RUNNING" {
            taskId := strings.Split(*task.TaskArn, "/")[2]
            taskDefinition := strings.Split(*task.TaskDefinitionArn, "/")[1]
            runtime := *task.Containers[0].RuntimeId
            tasks = append(tasks, fmt.Sprintf("%s | %s | %s", taskId, taskDefinition, runtime)) 
        }
    }
    if len(tasks) == 0 {
        log.Printf("There is no running ECS Task in the cluster %s", p.cluster)
        p.step <- "getCluster"
    }

    var taskInfo = strings.Split(SelectTask(tasks), " | ")
    p.task = taskInfo[0]
    p.runtime = taskInfo[2]
    p.step <- "runExecuteCommand"
}

func (p *Execs) runExecuteCommand() {

    p.command  = "/bin/sh"
    p.endpoint = fmt.Sprintf("https://ssm.%s.amazonaws.com", p.region)

    output, err := p.client.ExecuteCommand(context.TODO(), &ecs.ExecuteCommandInput{
        Cluster: &p.cluster,
        Task: &p.task,
        Interactive: true,
        Command: &p.command,
    })
    if err != nil {
        p.err <- err
        log.Fatalf("Failed to execute command, %v", err)
    }

    session, err := json.Marshal(output.Session)
    if err != nil {
        p.err <- err
        log.Fatalf("Json marshal for session is wrong, %v", err)
    }

    var target = fmt.Sprintf("ecs:%s_%s_%s", p.cluster, p.task, p.runtime)
    var ssmTarget = ssm.StartSessionInput{
        Target: &target,
    }
    targetJson, err := json.Marshal(ssmTarget)
    if err != nil {
        p.err <- err
        log.Fatalf("Json marshal for ssmTarget is wrong, %v", err)
    }

    cmd := exec.Command("session-manager-plugin", string(session), p.region, "StartSession", "", string(targetJson), p.endpoint)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGINT)
	go func() {
        <- s
	}()
	defer close(s)

    if err := cmd.Run(); err != nil {
        p.err <- err
		log.Fatal("Failed to run session-manager-plugin, is it installed?")
        log.Fatal(err)
	}

    p.quit <- true
}