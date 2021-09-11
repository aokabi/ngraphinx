package lib

import (
	"bufio"
	"net"
	"os"
	"strings"
	"time"

	ltsv "github.com/Songmu/go-ltsv"
)

const timeFormat = "02/Jan/2006:15:04:05 +0900"

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

type log struct {
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

func (l *log) GetEndPoint() string {
	return strings.Split(l.Req, " ")[1]
}

func GetNginxAccessLog(filepath string) ([]log, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)

	logs := make([]log, 0)
	for fileScanner.Scan() {
		var log log
		err := ltsv.Unmarshal(fileScanner.Bytes(), &log)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}
