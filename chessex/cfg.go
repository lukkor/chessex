package chessex

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type ServiceCfg struct {
	ScyllaCfg *ScyllaCfg `json:"scylla"`
	LoaderCfg *LoaderCfg `json:"loader"`
}

func NewDefaultServiceCfg() *ServiceCfg {
	return &ServiceCfg{
		ScyllaCfg: NewDefaultScyllaCfg(),
		LoaderCfg: NewDefaultLoaderCfg(),
	}
}

func LoadCfg(filePath string, serviceCfg *ServiceCfg) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", filePath, err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", filePath, err)
	}

	if err := json.Unmarshal(data, serviceCfg); err != nil {
		return fmt.Errorf("cannot parse json: %w", err)
	}

	return nil
}

func (sc *ServiceCfg) DumpCfg() string {
	json, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		panic(fmt.Errorf("cannot marshal json: %w", err))
	}

	return string(json)
}
