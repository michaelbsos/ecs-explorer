# ECS Explorer
This is a small utility to quickly connect to a container inside an ECS cluster.

## Usage
```sh
ecs-explorer --help
```
```
  -command string
    	the command to run inside the container (default "sh")
  -region string
    	the AWS region (default "ap-southeast-2")
```

## Building

Static build:
```sh
CGO_ENABLED=0 go build -ldflags "-s -w"
```