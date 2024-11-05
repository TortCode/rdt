package util

type Terminator struct {
	quit chan struct{} // quit signal is closed to terminate goroutine
	done chan struct{} // done signal is closed when goroutine is complete
}

func NewTerminator() *Terminator {
	return &Terminator{
		quit: make(chan struct{}),
		done: make(chan struct{}),
	}
}

func (term *Terminator) Quit() <-chan struct{} {
	return term.quit
}

func (term *Terminator) Done() {
	close(term.done)
}

func (term *Terminator) Terminate() {
	close(term.quit)
	<-term.done
}
