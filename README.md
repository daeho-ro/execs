# execs

[![Go Reference](https://pkg.go.dev/badge/github.com/daeho-ro/execs.svg)](https://pkg.go.dev/github.com/daeho-ro/execs)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/daeho-ro/execs)](.)
[![build](https://github.com/daeho-ro/execs/actions/workflows/go.yml/badge.svg)](https://github.com/daeho-ro/execs/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/daeho-ro/execs)](https://goreportcard.com/report/github.com/daeho-ro/execs)
[![CodeFactor](https://www.codefactor.io/repository/github/daeho-ro/execs/badge/main)](https://www.codefactor.io/repository/github/daeho-ro/execs/overview/main)
[![GitHub](https://img.shields.io/github/license/daeho-ro/execs)](https://github.com/daeho-ro/execs/blob/main/LICENSE)
[![Brew Version](https://img.shields.io/badge/dynamic/json.svg?url=https://raw.githubusercontent.com/draftbrew/homebrew-tap/master/Info/execs.json&query=$.version&label=homebrew)](https://github.com/draftbrew/homebrew-tap)
[![Choco version](https://img.shields.io/chocolatey/v/execs)](https://community.chocolatey.org/packages/execs)

## Introduction

**execs** is a program that helps you to access the ECS task interactively by using the `ssm` session-manager-plugin. It uses the ECS execute command API with the command `/bin/sh`. It also highly refer the similar and pre-existing program [ecsgo](https://github.com/tedsmitt/ecsgo) but uses the AWS SDK for GO v2. Since the motivation for developing the program is personal purpose to study about Golang, the program could be unstable.

## Installation
From the release, you can download the package directly.
### MacOS
You can use `brew` to install the package.
```bash
brew tap draftbrew/repo
brew install execs
```

### Windows
You can use `chocolatey` to install the package.
```bash
choco install execs
```
You also can use `scoop` to install the custom application as follows.
```bash
scoop install https://raw.githubusercontent.com/daeho-ro/execs-chocolatey-package/main/execs.json
```

## AWS Profile
In order to use `execs`, you need to pass the AWS credentials. As far as I know, the best way to do it is using `aws-vault`.
```bash
aws-vault exec <profile> -- execs
```
Please check the [aws-vault](https://github.com/99designs/aws-vault) page.

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
