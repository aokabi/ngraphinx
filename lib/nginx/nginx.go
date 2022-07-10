package nginx

import (
	"bufio"
	"errors"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	ltsv "github.com/Songmu/go-ltsv"
)

const timeFormat = "02/Jan/2006:15:04:05 -0700"

type logTime struct {
	time.Time
}

func (lt *logTime) UnmarshalText(t []byte) error {
	ti, err := time.ParseInLocation(timeFormat, string(t), time.UTC)
	if err != nil {
		return err
	}
	lt.Time = ti
	return nil
}

type Log struct {
	Time    *logTime
	Host    net.IP
	Req     string
	Status  int
	Size    int
	UA      string
	ReqTime float64
	AppTime *float64
	VHost   string
}

func (l *Log) GetEndPoint() (string, error) {
	reqs := strings.Split(l.Req, " ")
	if len(reqs) <= 1 {
		return "", errors.New("endpoint error")
	}
	req := reqs[1]

	// alpに習ってクエリパラメータは除去
	u, err := url.Parse(req)
	if err != nil {
		return "", err
	}
	return u.Path, nil
}

func (l *Log) GetMethod() string {
	return strings.Split(l.Req, " ")[0]
}

func GetNginxAccessLog(filepath string) ([]Log, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)

	logs := make([]Log, 0)
	for fileScanner.Scan() {
		var log Log
		err := ltsv.Unmarshal(fileScanner.Bytes(), &log)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}
