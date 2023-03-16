package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

var (
	cfg aws.Config
	svc *ecs.Client
)

// selectECSCluster fetches a list of ECS clusters and presents a list to the user
// the user then selects the desired cluster and it is returned
func selectECSCluster() (types.Cluster, error) {
	cluster := types.Cluster{}

	// List ECS Clusters
	resp, err := svc.ListClusters(context.TODO(), &ecs.ListClustersInput{})
	if err != nil {
		return cluster, fmt.Errorf("failed to list clusters: %s", err)
	}

	// Describe ECS Clusters
	r, err := svc.DescribeClusters(context.TODO(), &ecs.DescribeClustersInput{Clusters: resp.ClusterArns})
	if err != nil {
		return cluster, fmt.Errorf("failed to describe clusters: %s", err)
	}

	// Display selection to user
	for i, cluster := range r.Clusters {
		fmt.Printf("%d | %s\n", i, *cluster.ClusterName)
	}

	// Fetch selection from user
	var input string
	fmt.Printf("\nSelect an ECS Cluster: ")

	_, err = fmt.Scanln(&input)
	if err != nil {
		return cluster, fmt.Errorf("failed to get user input: %s", err)
	}

	selection, err := strconv.Atoi(input)
	if err != nil {
		return cluster, fmt.Errorf("selection was not a number: %s", err)
	}

	if selection < 0 || selection > len(r.Clusters) {
		log.Fatal("selection out of range")
	}

	return r.Clusters[selection], nil
}

// selectClusterTask fetches a list of ECS tasks and presents a list to the user
// the user then selects the desired task and it is returned
func selectClusterTask(cluster types.Cluster) (types.Task, error) {
	task := types.Task{}

	// List ECS Tasks
	resp, err := svc.ListTasks(context.TODO(), &ecs.ListTasksInput{
		Cluster: cluster.ClusterArn,
	})
	if err != nil {
		return task, fmt.Errorf("failed to list tasks: %s", err)
	}

	// Describe ECS Tasks
	r, err := svc.DescribeTasks(context.TODO(), &ecs.DescribeTasksInput{
		Cluster: cluster.ClusterArn,
		Tasks:   resp.TaskArns,
	})
	if err != nil {
		return task, fmt.Errorf("failed to describe tasks: %s", err)
	}

	// Display selection to user
	for i, task := range r.Tasks {
		fmt.Printf("%d | %s\n", i, *task.TaskArn)

		for _, container := range task.Containers {
			fmt.Printf("  * %s\n", *container.Name)
		}
	}

	// Fetch selection from user
	var input string
	fmt.Printf("\nSelect an ECS Task: ")

	_, err = fmt.Scanln(&input)
	if err != nil {
		return task, fmt.Errorf("failed to get user input: %s", err)
	}

	selection, err := strconv.Atoi(input)
	if err != nil {
		return task, fmt.Errorf("selection was not a number: %s", err)
	}

	if selection < 0 || selection > len(r.Tasks) {
		log.Fatal("selection out of range")
	}

	return r.Tasks[selection], nil
}

// selectTaskContainer fetches a list of ECS tasks and presents a list to the user
// the user then selects the desired task and it is returned
func selectTaskContainer(task types.Task) (types.Container, error) {
	container := types.Container{}

	if len(task.Containers) == 1 {
		return task.Containers[0], nil
	}

	// Display selection to user
	for i, c := range task.Containers {
		fmt.Printf("%d | %s\n", i, *c.Name)
	}

	// Fetch selection from user
	var input string
	fmt.Printf("\nSelect a Container: ")

	_, err := fmt.Scanln(&input)
	if err != nil {
		return container, fmt.Errorf("failed to get user input: %s", err)
	}

	selection, err := strconv.Atoi(input)
	if err != nil {
		return container, fmt.Errorf("selection was not a number: %s", err)
	}

	if selection < 0 || selection > len(task.Containers) {
		log.Fatal("selection out of range")
	}

	return task.Containers[selection], nil
}

func main() {
	command := flag.String("command", "sh", "the command to run inside the container")
	region := flag.String("region", "ap-southeast-2", "the AWS region")
	flag.Parse()

	var err error
	// Create AWS Config
	cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(*region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create ECS Client
	svc = ecs.NewFromConfig(cfg)

	// Get the ECS Cluster
	cluster, err := selectECSCluster()
	if err != nil {
		log.Fatalf("error selecting cluster: %s\n", err)
	}
	fmt.Println()

	// Get the Clusters Task
	task, err := selectClusterTask(cluster)
	if err != nil {
		log.Fatalf("error selecting task: %s\n", err)
	}
	fmt.Println()

	// Get the Tasks Container
	container, err := selectTaskContainer(task)
	if err != nil {
		log.Fatalf("error selecting container: %s\n", err)
	}
	fmt.Println()

	// Jump inside the container using aws cli tools
	cmd := exec.Command("aws", "ecs", "execute-command", "--cluster", *cluster.ClusterName, "--task", *task.TaskArn, "--container", *container.Name, "--interactive", "--command", *command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Fatalf("error while running aws command: %s\n", err)
	}
}
