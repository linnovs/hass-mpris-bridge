package bufferpool

import (
	"bytes"
	"sync"
)

var pool = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

func Get() *bytes.Buffer {
	return pool.Get().(*bytes.Buffer)
}

func Put(b *bytes.Buffer) {
	b.Reset()
	pool.Put(b)
}
