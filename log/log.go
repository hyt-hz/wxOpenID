package log

import (
	"github.com/cihub/seelog"
)

func Trace(format string, v ...interface{}) {
	seelog.Tracef(format, v...)
}

func Debug(format string, v ...interface{}) {
	seelog.Debugf(format, v...)
}

func Info(format string, v ...interface{}) {
	seelog.Infof(format, v...)
}

func Warning(format string, v ...interface{}) {
	seelog.Warnf(format, v...)
}

func Error(format string, v ...interface{}) {
	seelog.Errorf(format, v...)
}

func Critical(format string, v ...interface{}) {
	seelog.Criticalf(format, v...)
}

var default_conf string = `
<seelog type="sync">
    <outputs formatid="main">
        <console />
    </outputs>
    <formats>
        <format id="main" format="[%Date(2006-01-02 03:04:05.000000000 MST)] [%Level] %Msg%n"/>
        <format id="main2" format="%Date/%Time [%Level] %Msg%n"/>
        <format id="default" format="%Ns [%Level] %Msg%n"/>
    </formats>
</seelog>
`

func init() {
	logger, err := seelog.LoggerFromConfigAsString(default_conf)
	if err != nil {
		Warning("Parsing default config error,and use system default config. err:%v", err)
		return
	}
	seelog.ReplaceLogger(logger)
	return
}

func Start(conf_file string) {
	logger, err := seelog.LoggerFromConfigAsFile(conf_file)
	if err != nil {
		Warning("Parsing config file %v error,and use default config.err:%v", conf_file, err)
		return
	}
	seelog.ReplaceLogger(logger)
	Info("Start use log conf file %v", conf_file)
	return
}

func Flush() {
	seelog.Flush()
}
