package runtime

import (
	"os/exec"
	"sync"
)

type Runtime struct {
	wg sync.WaitGroup
}

func New() *Runtime {
	return &Runtime{}
}

// Wait blocks until all background tasks started via Go have completed.
func (r *Runtime) Wait() {
	r.wg.Wait()
}

func (r *Runtime) Go(fn func()) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		fn()
	}()
}

func (r *Runtime) HasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
