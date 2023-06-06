package utils

import "k8s.io/apimachinery/pkg/util/json"

func SwapTo(request, target interface{}) (err error) {
	dataByte, err := json.Marshal(request)
	if err != nil {
		return
	}

	err = json.Unmarshal(dataByte, target)
	return nil

}
