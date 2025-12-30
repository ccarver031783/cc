package whoami

import (
	"fmt"
	"os/exec"
	"strings"

	ufcli "github.com/urfave/cli/v2"
)

// WhoamiResult holds the result of the whoami command
type WhoamiResult struct {
	GitUser    string
	AWSUser    string
	K8sContext string
}

func getGitUser() (string, error) {
	nameCmd := exec.Command("git", "config", "user.name")
	nameOut, nameErr := nameCmd.Output()

	emailCmd := exec.Command("git", "config", "user.email")
	emailOut, emailErr := emailCmd.Output()

	// If both fail, return error
	if nameErr != nil && emailErr != nil {
		return "", nameErr
	}

	name := strings.TrimSpace(string(nameOut))
	email := strings.TrimSpace(string(emailOut))

	if name == "" && email == "" {
		return "", fmt.Errorf("git user name and email are not set")
	}

	return fmt.Sprintf("%s <%s>", name, email), nil
}

func getAWSUser() (string, error) {
	cmd := exec.Command("aws", "sts", "get-caller-identity", "--query", "Arn", "--output", "text")
	output, err := cmd.Output()
	if err != nil {
		return "", err // Return the error
	}
	return strings.TrimSpace(string(output)), nil
}

func getK8sContext() (string, error) {
	cmd := exec.Command("kubectl", "config", "current-context")
	output, err := cmd.Output()
	if err != nil {
		return "", err // Return the error
	}
	return strings.TrimSpace(string(output)), nil
}

func Whoami() (WhoamiResult, error) {
	result := WhoamiResult{}
	var errors []string

	if gitUser, err := getGitUser(); err != nil {
		errors = append(errors, fmt.Sprintf("git: %v", err))
		result.GitUser = "(unavailable)"
	} else {
		result.GitUser = gitUser
	}

	if awsUser, err := getAWSUser(); err != nil {
		errors = append(errors, fmt.Sprintf("aws: %v", err))
		result.AWSUser = "(unavailable)"
	} else {
		result.AWSUser = awsUser
	}

	if k8sContext, err := getK8sContext(); err != nil {
		errors = append(errors, fmt.Sprintf("k8s: %v", err))
		result.K8sContext = "(unavailable)"
	} else {
		result.K8sContext = k8sContext
	}

	// Return result always, but include errors if any
	if len(errors) > 0 {
		return result, fmt.Errorf("partial failures: %s", strings.Join(errors, "; "))
	}
	return result, nil
}

func NewWhoamiCmd() *ufcli.Command {
	return &ufcli.Command{
		Name:  "whoami",
		Usage: "Display current Git user, AWS identity, and K8s context in one command",
		Action: func(c *ufcli.Context) error {
			result, err := Whoami()

			fmt.Printf("Git:     %s\n", result.GitUser)
			fmt.Printf("AWS:     %s\n", result.AWSUser)
			fmt.Printf("K8s:     %s\n", result.K8sContext)

			//Return the error if there is one
			return err
		},
	}
}
