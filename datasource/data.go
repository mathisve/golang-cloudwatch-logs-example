package datasource

import (
	"github.com/zhexuany/wordGenerator"
	"sync"
	"time"
)

func GenerateData(arr *[]string, lock *sync.Mutex) {
	for {
		lock.Lock()
		*arr = append(*arr, wordGenerator.GetWord(10))
		lock.Unlock()

		time.Sleep(time.Second * 2)
	}
}