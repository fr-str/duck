package docker

import (
	"bufio"
	"bytes"
	"context"
	dcli "docker-project/docker/client"
	log "docker-project/logger"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
)

type Log struct {
	Container string
	Message   string
	Timestamp int64
}

var (
	reg             = `\x1b\[[0-9]{1,2}(?:m|;[0-9];[0-9]{1,3}m)`
	regex           = regexp.MustCompile(reg)
	ErrContNotExist = errors.New("container does not exist")
)

func GetLogs(contName string, amount int, since, until int64, follow bool) ([]Log, io.ReadCloser, error) {
	cont, ok := Containers.GetFull(contName)
	if !ok {
		return nil, nil, ErrContNotExist
	}
	log.Debugw("Reading logs for "+contName, "amount", amount, "since", since, "until", until, "follow", follow)

	tsince := time.Unix(0, since).Format(time.RFC3339Nano)
	tuntil := time.Unix(0, until).Format(time.RFC3339Nano)
	if since == 0 {
		tsince = "0"
	}
	if until == 0 {
		tuntil = "0"
	}

	// TODO better way to navigate thru logs
	// maybe just get container logfile??
	// TODO add parser to separate out timestamp,level and message
	untilAmount := amount
	if until != 0 {
		untilAmount = 1000
	}

	rc, err := dcli.Cli.ContainerLogs(context.TODO(), contName, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Since:      tsince,
		Until:      tuntil,
		Timestamps: true,
		Follow:     follow,
		Tail:       strconv.Itoa(untilAmount),
		Details:    false,
	})
	if err != nil {
		log.Error(err)
		return nil, nil, err
	}

	if follow {
		return nil, rc, nil
	}

	var r io.Reader = rc

	if !cont.Tty {
		buff := bytes.NewBuffer(make([]byte, 0))
		r = buff
		stdcopy.StdCopy(buff, buff, rc)
	}

	var i int
	var line string
	logs := []Log{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line = sc.Text()
		//docker timestamp
		t := line[:30]
		msg := line[30:]

		if len(msg) != 0 {
			msg = msg[1:]
		}
		ti, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(t))
		if err != nil {
			log.Error(err)
		}

		logs = append(logs, Log{
			Timestamp: ti.UnixNano(),
			Message:   CutTimestamp(msg),
			Container: contName,
		})
		i++
	}
	if until != 0 && i-amount >= len(logs) {
		logs = logs[i-amount:]
	}

	log.Debug("log num", i)
	return logs, nil, nil
}

// it's simple, but it works... sometimes
func CutTimestamp(line string) string {
	line = regex.ReplaceAllString(line, "")
	var i int
	for i = 0; i < len(line); i++ {
		r := line[i]
		switch {
		case r >= '0' && r <= '9':
			continue
		case r == '-', r == 'T', r == ':', r == '.', r == 'Z', r == ' ', r == '+', r == '/': //, r == '[', r == ']':
			continue
		default:
			return line[i:]
		}
	}
	return line
}
