package slf4goipfs

import (
	"encoding/json"

	ipfslog "github.com/ipfs/go-log/v2"
	"github.com/libs4go/slf4go"
)

func init() {
	ipfslog.SetupLogging(ipfslog.Config{})
	ipfslog.SetAllLoggers(ipfslog.LevelDebug)

	reader := ipfslog.NewPipeReader()

	go redirectLoop(reader)
}

type loggerJSON struct {
	Level   string `json:"level"`
	TS      string `json:"ts"`
	Logger  string `json:"logger"`
	Message string `json:"msg"`
	Error   string `json:"error"`
}

var errlog = slf4go.Get("ipfslog")

type logF func(message string, args ...interface{})

func getLogF(level string, logger slf4go.Logger) logF {
	l, err := ipfslog.LevelFromString(level)

	if err != nil {
		errlog.E("decode ipfs log level {@level} error {@error}", level, err)

		return logger.I
	}

	switch l {
	case ipfslog.LevelDebug:
		return logger.D
	case ipfslog.LevelInfo:
		return logger.I
	case ipfslog.LevelWarn:
		return logger.W
	case ipfslog.LevelError:
		return logger.E
	default:
		return logger.E
	}
}

func redirectLoop(reader *ipfslog.PipeReader) {

	logs := make(map[string]slf4go.Logger)

	for {
		decoder := json.NewDecoder(reader)

		// var message interface{}

		// err := decoder.Decode(&message)

		// if err != nil {
		// 	errlog.E("decode ipfs log error {@error}", err)
		// 	continue
		// }

		// buff, _ := json.Marshal(message)

		// errlog.I(string(buff))

		var logEntry loggerJSON

		err := decoder.Decode(&logEntry)

		if err != nil {
			errlog.E("decode ipfs log error {@error}", err)
			continue
		}

		log, ok := logs[logEntry.Logger]

		if !ok {
			log = slf4go.Get(logEntry.Logger)

			logs[logEntry.Logger] = log
		}

		f := getLogF(logEntry.Level, log)

		if logEntry.Error != "" {
			f(logEntry.Message)
			log.E(logEntry.Error)
		} else {
			f(logEntry.Message)
		}

	}
}
