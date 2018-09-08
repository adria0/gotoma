package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
)

type KVStorage struct {
	sync.Mutex
	values   map[string]string
	filename string
}

func NewKVStorage(filename string) (*KVStorage, error) {

	values := make(map[string]string)

	content, err := ioutil.ReadFile(filename)
	if err == nil {
		for _, line := range strings.Split(string(content), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") || len(trimmed) == 0 {
				continue
			}
			kv := strings.SplitN(line, "=", 2)
			values[kv[0]] = kv[1]
		}
	}
	return &KVStorage{
		values:   values,
		filename: filename,
	}, nil
}

func (kv *KVStorage) Get(k string) (string, bool) {
	kv.Lock()
	value, ok := kv.values[k]
	kv.Unlock()
	return value, ok
}

func (kv *KVStorage) Put(k, v string) error {
	kv.Lock()
	defer kv.Unlock()

	kv.values[k] = v

	var buffer bytes.Buffer
	for k, v = range kv.values {
		buffer.WriteString(fmt.Sprintf("%v=%v\n", k, v))
	}

	return ioutil.WriteFile(kv.filename, buffer.Bytes(), 0744)
}
