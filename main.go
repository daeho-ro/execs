package main

func main() {

    p := &Execs{
        region : "ap-northeast-2",
        endpoint : "https://ssm.ap-northeast-2.amazonaws.com",
    }
    
    GetEcsSession(p)
    GetCluster(p)
    GetTask(p)
    RunExecuteCommand(p)
}
