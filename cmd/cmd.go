package cmd

import (
	"os/exec"
	"fmt"
)

func Execute(command, workingDirectory string) (string, error) {
	c := exec.Command("sh", "-c", command)
	c.Dir = workingDirectory
	output, err := c.Output()
	if err != nil {
		fmt.Println(workingDirectory)
		fmt.Println(fmt.Sprintf("Failed to execute %s: %s\n", command, err))
		return "", err
	}

	fmt.Println(workingDirectory)
	fmt.Println(command)
	fmt.Println(fmt.Sprintf("%s", output))

	return fmt.Sprintf("%s", output), nil
}

func RemoveFolder(folder string) error {
	_, err := Execute(fmt.Sprintf("rm -rf %s", folder), "/")
	return err
}
