package flashduty

import (
	"fmt"
	"net/url"
	"time"

	"github.com/ccfos/nightingale/v6/center/cconf"
	"github.com/ccfos/nightingale/v6/pkg/poster"

	"github.com/toolkits/pkg/logger"
)

var (
	Api     string
	Timeout time.Duration
)

func Init(fdConf cconf.FlashDuty) {
	Api = fdConf.Api
	if fdConf.Timeout == 0 {
		Timeout = 5 * time.Second
	} else {
		Timeout = fdConf.Timeout * time.Millisecond
	}
}

func PostFlashDuty(path string, appKey string, body interface{}) error {
	urlParams := url.Values{}
	urlParams.Add("app_key", appKey)
	var url string
	if Api != "" {
		url = fmt.Sprintf("%s%s?%s", Api, path, urlParams.Encode())
	} else {
		url = fmt.Sprintf("%s%s?%s", "https://api.flashcat.cloud", path, urlParams.Encode())
	}
	response, code, err := poster.PostJSON(url, Timeout, body)
	logger.Infof("exec PostFlashDuty: url=%s, body=%v; response=%s, code=%d", url, body, response, code)
	return err
}
