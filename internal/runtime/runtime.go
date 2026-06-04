package runtime

import (
	"context"
	"os/exec"
	"sync"
)

type Runtime struct {
	wg sync.WaitGroup

	ctx    context.Context
	cancel context.CancelFunc
}

func New() *Runtime {
	ctx, cancel := context.WithCancel(context.Background())
	return &Runtime{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Wait blocks until all background tasks started via Go have completed.
func (r *Runtime) Wait() {
	r.wg.Wait()
}

// Shutdown cancels all commands started via Command and signals Done.
func (r *Runtime) Shutdown() {
	r.cancel()
}

func (r *Runtime) Done() <-chan struct{} {
	return r.ctx.Done()
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

// Command returns an exec.Cmd bound to the runtime lifecycle.
// The command is automatically cancelled when Shutdown is called.
func (r *Runtime) Command(name string, args ...string) *exec.Cmd {
	return exec.CommandContext(r.ctx, name, args...)
}
