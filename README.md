# execs

[![Go Reference](https://pkg.go.dev/badge/github.com/daeho-ro/execs.svg)](https://pkg.go.dev/github.com/daeho-ro/execs)
![build](https://github.com/daeho-ro/execs/actions/workflows/go.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/daeho-ro/execs)](https://goreportcard.com/report/github.com/daeho-ro/execs)
[![CodeFactor](https://www.codefactor.io/repository/github/daeho-ro/execs/badge/main)](https://www.codefactor.io/repository/github/daeho-ro/execs/overview/main)
[![Hits](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Fdaeho-ro%2Fexecs&count_bg=%2379C83D&title_bg=%23555555&icon=&icon_color=%23E7E7E7&title=hits&edge_flat=false)](https://hits.seeyoufarm.com)
![GitHub](https://img.shields.io/github/license/daeho-ro/execs)

## Introduction

**execs** is a program that helps you to access the ECS task interactively by using the `ssm` session-manager-plugin. It uses the ECS execute command API with the command `/bin/sh`. It also highly refer the similar and pre-existing program [ecsgo](https://github.com/tedsmitt/ecsgo) but uses the AWS SDK for GO v2. Since the motivation for developing the program is personal purpose to study about Golang, the program could be unstable.

To use the program:
- download the binary that is fit to your OS and CPU architecture
- place it to the directory in your PATH or add the PATH to the program
- just run `execs`

I use the program with the AWS environment variables that could be obtained from the AWS SSO when you use the organizational and consolidated accounts or AWS STS when you use the assume role. Once you get the environment variables, export it as follows:
```
export AWS_ACCESS_KEY_ID="ASIAXXXXXXXXXXXXXXXX"
export AWS_SECRET_ACCESS_KEY="<Secrets>"
export AWS_SESSION_TOKEN="<Long_Token>"
```
In addition, you can also use the default profile of the AWS credentials (~/.aws/credentials, ~/.aws/config) which can be set by manual or the `configure` command by the AWS CLI. The other profils are not yet supported.

## To-do
- [ ] China support
- [ ] Profile support

## Required permissions

This program use the AWS IAM permissions as follows:
```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "execs-iam-permissions",
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeRegions",
                "ecs:DescribeTasks",
                "ecs:ExecuteCommand",
                "ecs:ListClusters",
                "ecs:ListTasks"
            ],
            "Resource": "*"
        }
    ]
}
```
