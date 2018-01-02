package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/dshearer/jobber/common"
	"os"
	"os/user"
	"strings"
	"text/tabwriter"
	"time"
)

type ListRespRec struct {
	usr  *user.User
	resp *common.ListJobsCmdResp
}

func formatTime(t *time.Time) string {
	if t == nil {
		return "none"
	} else {
		tmp := t.Local()
		return tmp.Format("Jan _2 15:04:05 -0700 MST")
	}
}

func doListCmd_allUsers() int {
	// get all users
	users, err := common.AllUsersWithSockets()
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to get all users: %v\n", err,
		)
		return 1
	}

	// send cmd
	var responses []ListRespRec
	for _, usr := range users {
		var resp common.ListJobsCmdResp
		err = CallDaemon(
			"NewIpcService.ListJobs",
			common.ListJobsCmd{},
			&resp,
			usr,
			true,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"Failed to list jobs for %v: %v\n", usr.Name, err)
			continue
		}
		rec := ListRespRec{usr: usr, resp: &resp}
		responses = append(responses, rec)
	}

	// make table header
	var buffer bytes.Buffer
	var writer *tabwriter.Writer = tabwriter.NewWriter(&buffer,
		5, 0, 2, ' ', 0)
	headers := [...]string{
		"NAME",
		"STATUS",
		"SEC/MIN/HR/MDAY/MTH/WDAY",
		"NEXT RUN TIME",
		"NOTIFY ON SUCCESS",
		"NOTIFY ON ERR",
		"NOTIFY ON FAIL",
		"ERR HANDLER",
		"USER",
	}
	fmt.Fprintf(writer, "%v\n", strings.Join(headers[:], "\t"))

	// make table rows
	var rows []string
	for _, respRec := range responses {
		var userName string
		if len(respRec.usr.Name) > 0 {
			userName = respRec.usr.Name
		} else {
			userName = respRec.usr.Username
		}
		for _, j := range respRec.resp.Jobs {
			var s string
			s = fmt.Sprintf(
				"%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v",
				j.Name,
				j.Status,
				j.Schedule,
				formatTime(j.NextRunTime),
				j.NotifyOnSuccess,
				j.NotifyOnErr,
				j.NotifyOnFail,
				j.ErrHandler,
				userName)
			rows = append(rows, s)
		}
	}
	fmt.Fprintf(writer, "%v", strings.Join(rows, "\n"))
	writer.Flush()
	fmt.Printf("%v\n", buffer.String())
	return 0
}

func doListCmd_currUser() int {
	// get current user
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to get current user: %v\n", err,
		)
		return 1
	}

	// send cmd
	var resp common.ListJobsCmdResp
	err = CallDaemon(
		"NewIpcService.ListJobs",
		common.ListJobsCmd{},
		&resp,
		usr,
		true,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// make table header
	var buffer bytes.Buffer
	var writer *tabwriter.Writer = tabwriter.NewWriter(&buffer,
		5, 0, 2, ' ', 0)
	headers := [...]string{
		"NAME",
		"STATUS",
		"SEC/MIN/HR/MDAY/MTH/WDAY",
		"NEXT RUN TIME",
		"NOTIFY ON ERR",
		"NOTIFY ON FAIL",
		"ERR HANDLER",
	}
	fmt.Fprintf(writer, "%v\n", strings.Join(headers[:], "\t"))

	// handle response
	strs := make([]string, 0, len(resp.Jobs))
	for _, j := range resp.Jobs {
		var s string
		if usr != nil {
			s = fmt.Sprintf("%v\t", usr.Name)
		}
		s = fmt.Sprintf(
			"%v\t%v\t%v\t%v\t%v\t%v\t%v",
			j.Name,
			j.Status,
			j.Schedule,
			formatTime(j.NextRunTime),
			j.NotifyOnErr,
			j.NotifyOnFail,
			j.ErrHandler)
		strs = append(strs, s)
	}
	fmt.Fprintf(writer, "%v", strings.Join(strs, "\n"))
	writer.Flush()
	fmt.Printf("%v\n", buffer.String())

	return 0
}

func doListCmd(args []string) int {
	// parse flags
	flagSet := flag.NewFlagSet(ListCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(ListCmdStr, "", flagSet)
	var help_p = flagSet.Bool("h", false, "help")
	var allUsers_p = flagSet.Bool("a", false, "all-users")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		return 0
	}

	if *allUsers_p {
		return doListCmd_allUsers()
	} else {
		return doListCmd_currUser()
	}
}