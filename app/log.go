package app

import log "github.com/cihub/seelog"

func InitLog(logConfigFile string)  {
	defer log.Flush()
	filename, err := GetAbsPath(logConfigFile)
	ThrowErr(err)
	logger, err := log.LoggerFromConfigAsFile(filename)
	ThrowErr(err)
	log.UseLogger(logger)
}
