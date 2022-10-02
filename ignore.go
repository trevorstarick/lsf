package lsf

import (
	"sync"
)

var ignoreList = sync.Map{}

func ResetIgnoreList() {
	ignoreList = sync.Map{}
}

func AddToIgnoreList(s string) {
	ignoreList.Store(s, struct{}{})
}

func inIgnoreList(p []byte) bool {
	_, ok := ignoreList.Load(string(p))

	return ok
}
