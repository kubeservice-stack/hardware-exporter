package hosts

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"hardware_exporter/lib/config"
)

type ServerMap map[string]string

var (
	serverInfo = make(ServerMap)
)

func ReadConfigIni(hostsConfigPath string, logger log.Logger) *ServerMap {
	cfg, err := config.ReadDefault(hostsConfigPath) //读取配置文件，并返回其Config
	if err != nil {
		level.Error(logger).Log("msg", "Fail to find", "config", hostsConfigPath, "err", err)
	}

	if cfg.HasSection("servers") { //判断配置文件中是否有section（一级标签）
		options, err := cfg.SectionOptions("servers") //获取一级标签的所有子标签options（只有标签没有值）
		if err == nil {
			for _, v := range options {
				optionValue, err := cfg.String("servers", v) //根据一级标签section和option获取对应的值
				if err == nil {
					serverInfo[v] = optionValue
				}
			}
		}
	}
	return &serverInfo

}
