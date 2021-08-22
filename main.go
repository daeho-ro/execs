package main

import "strings"

func main() {

    var region   = "ap-northeast-2"
    var ecs      = GetEcsSession(region)
    var clusters = GetClusters(ecs)
    var cluster  = SelectCluster(clusters)
    var tasks    = GetTasks(ecs, cluster)
    var task     = SelectTask(tasks)

    var taskId = strings.Split(task, " | ")[0]
    var runtime = strings.Split(task, " | ")[2]

    RunExecuteCommand(ecs, cluster, taskId, runtime)
}
