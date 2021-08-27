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

type execs struct {
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

// Start ...
func Start() {

	p := &execs{
		step:   make(chan string, 1),
		err:    make(chan error, 1),
		quit:   make(chan bool, 1),
		region: "ap-northeast-2",
	}

	p.loop()
}

func (p *execs) loop() {

	go func() {
		for {
			select {
			case step := <-p.step:
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
			case err := <-p.err:
				log.Fatal(err)
				os.Exit(1)
			}
		}
	}()

	p.step <- "getEcsSession"

	<-p.quit
}

func (p *execs) getEcsSession() {

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(p.region))
	if err != nil {
		log.Fatalf("Unable to load SDK config, %v", err)
		p.err <- err
	}

	p.client = *ecs.NewFromConfig(cfg)
	p.step <- "getCluster"
}

func (p *execs) getCluster() {

	var clusters []string
	pager := ecs.NewListClustersPaginator(&p.client, &ecs.ListClustersInput{})

	for pager.HasMorePages() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			log.Fatalf("failed to get a page, %v", err)
			p.err <- err
		}
		for _, arn := range page.ClusterArns {
			clusterName := strings.Split(arn, "/")[1]
			clusters = append(clusters, clusterName)
		}
	}

	if len(clusters) == 0 {
		log.Printf("There is no ECS clusters in the %s region", p.region)
		p.quit <- true
	}

	p.cluster = selectCluster(clusters)
	p.step <- "getTask"
}

func (p *execs) getTask() {

	var taskArns []string
	pager := ecs.NewListTasksPaginator(&p.client, &ecs.ListTasksInput{
		Cluster: &p.cluster,
	})

	for pager.HasMorePages() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			log.Fatalf("failed to get a page, %v", err)
			p.err <- err
		}
		taskArns = append(taskArns, page.TaskArns...)
	}

	var tasks []string
	for i := 0; i < len(taskArns)/100+1; i++ {

		var slice []string
		if i == 0 {
			slice = taskArns[0 : len(taskArns)%100]
		} else {
			slice = taskArns[i*100 : (i+1)*100]
		}

		listTaskDetails, err := p.client.DescribeTasks(context.TODO(), &ecs.DescribeTasksInput{
			Tasks:   slice,
			Cluster: &p.cluster,
		})
		if err != nil {
			p.err <- err
			log.Fatalf("Failed to describe ecs tasks, %v", err)
		}

		for _, task := range listTaskDetails.Tasks {
			if *task.LastStatus == "RUNNING" {
				taskID := strings.Split(*task.TaskArn, "/")[2]
				taskDefinition := strings.Split(*task.TaskDefinitionArn, "/")[1]
				runtime := strings.Split(*task.Containers[0].RuntimeId, "-")[1]
				tasks = append(tasks, fmt.Sprintf("%s | %s | %s", taskID, taskDefinition, runtime))
			}
		}
		if len(tasks) == 0 {
			log.Printf("There is no running ECS Task in the cluster %s", p.cluster)
			p.step <- "getCluster"
		}
	}

	var taskInfo = strings.Split(selectTask(tasks), " | ")
	p.task = taskInfo[0]
	p.runtime = fmt.Sprintf("%s-%s", p.task, taskInfo[2])
	p.step <- "runExecuteCommand"
}

func (p *execs) runExecuteCommand() {

	p.command = "/bin/sh"
	p.endpoint = fmt.Sprintf("https://ssm.%s.amazonaws.com", p.region)

	output, err := p.client.ExecuteCommand(context.TODO(), &ecs.ExecuteCommandInput{
		Cluster:     &p.cluster,
		Task:        &p.task,
		Interactive: true,
		Command:     &p.command,
	})
	if err != nil {
		p.err <- err
		log.Fatalf("Failed to execute command, %v", err)
	}

	session, err := json.Marshal(output.Session)
	if err != nil {
		log.Fatalf("Json marshal for session is wrong, %v", err)
		p.err <- err
	}

	var target = fmt.Sprintf("ecs:%s_%s_%s", p.cluster, p.task, p.runtime)
	var ssmTarget = ssm.StartSessionInput{
		Target: &target,
	}
	targetJSON, err := json.Marshal(ssmTarget)
	if err != nil {
		log.Fatalf("Json marshal for ssmTarget is wrong, %v", err)
		p.err <- err
	}

	cmd := exec.Command("session-manager-plugin", string(session), p.region, "StartSession", "", string(targetJSON), p.endpoint)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGINT)
	go func() {
		<-s
	}()
	defer close(s)

	if err := cmd.Run(); err != nil {
		log.Fatal("Failed to run session-manager-plugin, is it installed?")
		p.err <- err
	}

	p.quit <- true
}
