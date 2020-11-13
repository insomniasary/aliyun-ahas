package tools

import (
	"fmt"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/satori/go.uuid"
)

func GetUUID() string {
	rid, err := uuid.NewV4()
	if err != nil {
		return fmt.Sprintf("BAD_GENERATED_ID_%d", util.CurrentTimeMillis())
	}
	requestId := fmt.Sprintf("%s", rid)
	return requestId
}
