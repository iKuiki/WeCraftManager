package mcgate

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
)

// Advancement 进度
type Advancement struct {
	Advancement        string
	InGameDescription  string
	Parent             string
	ActualRequirements string
	InternalID         string
}

// loadAdvancement 加载进度表
func (mgt *MCGate) loadAdvancement(filepath string) error {
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return errors.WithStack(err)
	}
	var Advancements []Advancement
	err = json.Unmarshal(file, &Advancements)
	if err != nil {
		return errors.WithStack(err)
	}
	mgt.advancementMap = make(map[string]Advancement)
	for _, a := range Advancements {
		mgt.advancementMap[a.InternalID] = a
	}
	return nil
}
