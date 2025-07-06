package ecs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type execs struct {
	client      ecs.Client
	step        chan string
	err         chan error
	quit        chan bool
	cluster     string
	task        string
	runtime     string
	container   string
	region      string
	endpoint    string
	command     string
	taskDetails types.Task
}

// Start : Let's Go!
func Start() {

	region := os.Getenv("AWS_REGION")
	if len(region) == 0 {
		region = "ap-northeast-2"
	}

	p := &execs{
		step:   make(chan string, 1),
		err:    make(chan error, 1),
		quit:   make(chan bool, 1),
		region: region,
	}

	p.loop()
}

func (p *execs) loop() {

	go func() {
		for {
			select {
			case step := <-p.step:
				switch step {
				case "getRegion":
					p.getRegion()
				case "getEcsSession":
					p.getEcsSession()
				case "getCluster":
					p.getCluster()
				case "getTask":
					p.getTask()
				case "getContainer":
					p.getContainer()
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

	p.step <- "getRegion"

	<-p.quit
}

func (p *execs) getRegion() {

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(p.region))
	if err != nil {
		p.err <- fmt.Errorf("Unable to load SDK config: %v", err)
		return
	}

	ec2Client := *ec2.NewFromConfig(cfg)
	regionNames, err := ec2Client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		p.err <- fmt.Errorf("Failed to get the regions: %v", err)
		return
	}

	var regions []string
	for _, list := range regionNames.Regions {
		regions = append(regions, *list.RegionName)
	}
	sort.Strings(regions)

	p.region = selectRegion(regions, p.region)
	p.step <- "getEcsSession"
}

func (p *execs) getEcsSession() {

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(p.region))
	if err != nil {
		p.err <- fmt.Errorf("Unable to load SDK config: %v", err)
		return
	}

	p.client = *ecs.NewFromConfig(cfg)
	p.step <- "getCluster"
}

func (p *execs) getCluster() {

	clusters := []string{".."}
	pager := ecs.NewListClustersPaginator(&p.client, &ecs.ListClustersInput{})

	for pager.HasMorePages() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			p.err <- fmt.Errorf("Failed to get a page of the ECS clusters: %v", err)
			return
		}
		for _, arn := range page.ClusterArns {
			clusterName := strings.Split(arn, "/")[1]
			clusters = append(clusters, clusterName)
		}
	}

	if len(clusters) == 1 {
		log.Printf("There is no ECS clusters in the %s region", p.region)
		p.step <- "getRegion"
		return
	}

	selected := selectCluster(clusters)
	if selected == ".." {
		p.step <- "getRegion"
		return
	}
	p.cluster = selected
	p.step <- "getTask"
}

func (p *execs) getTask() {

	var taskArns []string
	pager := ecs.NewListTasksPaginator(&p.client, &ecs.ListTasksInput{
		Cluster:       &p.cluster,
		DesiredStatus: types.DesiredStatusRunning,
	})

	for pager.HasMorePages() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			p.err <- fmt.Errorf("Failed to get a page of the ECS tasks: %v", err)
			return
		}
		taskArns = append(taskArns, page.TaskArns...)
	}
	if len(taskArns) == 0 {
		log.Printf("There is no ECS Task in the cluster %s", p.cluster)
		p.step <- "getCluster"
		return
	}

	tasks := []string{".."}
	taskMap := make(map[string]types.Task)
	for i := 0; i < len(taskArns)/100+1; i++ {

		var slice []string
		var tail int = len(taskArns) % 100
		if i == 0 {
			slice = taskArns[0:tail]
		} else {
			slice = taskArns[tail+(i-1)*100 : tail+i*100]
		}

		listTaskDetails, err := p.client.DescribeTasks(context.TODO(), &ecs.DescribeTasksInput{
			Tasks:   slice,
			Cluster: &p.cluster,
		})
		if err != nil {
			p.err <- fmt.Errorf("Failed to describe ecs tasks: %v", err)
			return
		}

		for _, task := range listTaskDetails.Tasks {
			createdDt := task.CreatedAt.Format("2006-01-02 15:04:05")
			taskId := strings.Split(*task.TaskArn, "/")[2]
			taskDefinition := strings.Split(*task.TaskDefinitionArn, "/")[1]
			taskDisplay := fmt.Sprintf("%s | %-43s | %s", createdDt, taskId, taskDefinition)
			tasks = append(tasks, taskDisplay)
			taskMap[taskDisplay] = task
		}
	}
	if len(tasks) == 1 {
		log.Printf("There is no running ECS Task in the cluster %s", p.cluster)
		p.step <- "getCluster"
		return
	}

	selected := selectTask(tasks)
	if selected == ".." {
		p.step <- "getCluster"
		return
	}
	
	p.taskDetails = taskMap[selected]
	p.task = strings.Split(*p.taskDetails.TaskArn, "/")[2]
	p.step <- "getContainer"
}

func (p *execs) getContainer() {

	var runningContainers []string
	containerMap := make(map[string]types.Container)

	for _, container := range p.taskDetails.Containers {
		if container.RuntimeId != nil {
			containerName := *container.Name
			runningContainers = append(runningContainers, containerName)
			containerMap[containerName] = container
		}
	}

	if len(runningContainers) == 0 {
		log.Printf("There is no running container in the task %s", p.task)
		p.step <- "getTask"
		return
	}

	if len(runningContainers) == 1 {
		p.container = runningContainers[0]
		p.runtime = *containerMap[p.container].RuntimeId
	} else {
		containers := []string{".."}
		containers = append(containers, runningContainers...)
		
		selected := selectContainer(containers)
		if selected == ".." {
			p.step <- "getTask"
			return
		}
		
		p.container = selected
		p.runtime = *containerMap[selected].RuntimeId
	}

	p.step <- "runExecuteCommand"
}

func (p *execs) runExecuteCommand() {

	p.command = "/bin/sh"
	p.endpoint = fmt.Sprintf("https://ssm.%s.amazonaws.com", p.region)

	output, err := p.client.ExecuteCommand(context.TODO(), &ecs.ExecuteCommandInput{
		Cluster:     &p.cluster,
		Task:        &p.task,
		Container:   &p.container,
		Interactive: true,
		Command:     &p.command,
	})
	if err != nil {
		p.err <- fmt.Errorf("Failed to execute command: %v", err)
		return
	}

	session, err := json.Marshal(output.Session)
	if err != nil {
		p.err <- fmt.Errorf("Json marshal for session is wrong: %v", err)
		return
	}

	var target = fmt.Sprintf("ecs:%s_%s_%s", p.cluster, p.task, p.runtime)
	var ssmTarget = ssm.StartSessionInput{
		Target: &target,
	}
	targetJSON, err := json.Marshal(ssmTarget)
	if err != nil {
		p.err <- fmt.Errorf("Json marshal for ssmTarget is wrong: %v", err)
		return
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
		p.err <- fmt.Errorf("Failed to run session-manager-plugin, is it installed?: %v", err)
		return
	}

	p.quit <- true
}
