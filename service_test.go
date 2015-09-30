package service_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"../service"
)

func TestRespectsSignals(t *testing.T) {
	out, e := RunHelperProcess("TestHelperProcess", func(process *os.Process) {
		time.Sleep(100 * time.Millisecond)
		exec.Command("kill", "-INT", strconv.Itoa(process.Pid)).Run()
	})
	if e != nil {
		t.Errorf("Did not exit cleanly: %v", e)
	}
	fmt.Println(out)
	if !strings.HasPrefix(out, "loading!\nstarting!\nLooping!") {
		t.Error("Bad output for service")
	}
	if !strings.Contains(out, "Exiting cleanly") {
		t.Error("Process did not exit cleanly")
	}
}

func TestServesRequests(t *testing.T) {

}

func RunHelperProcess(testName string, f func(*os.Process)) (string, error) {
	var b bytes.Buffer

	cmd := exec.Command(os.Args[0], "-test.run="+testName)
	cmd.Env = []string{"HELPER_PROCESS=1"}
	cmd.Stdout = &b
	cmd.Stderr = &b
	e := cmd.Start()
	if e != nil {
		return "", e
	}
	if f != nil {
		f(cmd.Process)
	}
	e = cmd.Wait()
	return b.String(), e
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("HELPER_PROCESS") == "1" {
		defer os.Exit(0)
		serv := service.NewService()
		fmt.Println("loading!")
		serv.Start(func() {
			fmt.Println("starting!")
			serv.Timer(50*time.Millisecond, func() {
				fmt.Println("Looping!")
			})
		})
		serv.Wait()
		fmt.Println("Exiting cleanly")
	}
}
