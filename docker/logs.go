package docker

import (
	"bufio"
	"bytes"
	"context"
	log "docker-project/logger"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/moby/moby/pkg/stdcopy"
)

type Log struct {
	Timestamp int64
	Message   string
}

var (
	ErrContNotExist = errors.New("container does not exist")
)

func GetLogs(contName string, amount int, since, until int64, follow bool) ([]Log, io.ReadCloser, error) {
	if !ContainerMap.Exists(contName) {
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
	untilAmount := amount
	if until != 0 {
		untilAmount = 1000
	}

	rc, err := dcli.ContainerLogs(context.TODO(), contName, types.ContainerLogsOptions{
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

	if !ContainerMap.Get(contName).Tty {
		buff := bytes.NewBuffer(make([]byte, 0))
		r = buff
		stdcopy.StdCopy(buff, buff, rc)
	}

	var i int
	var line string
	logs := []Log{}

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line = string(sc.Bytes())
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
			Message:   msg,
		})
		i++
	}
	if until != 0 && i-amount >= len(logs) {
		logs = logs[i-amount:]
	}

	log.Debug("log num", i)
	return logs, nil, nil

}
