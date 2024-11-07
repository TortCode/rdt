package util

// Terminator provides mechanisms to signal
// and wait for termination of a goroutine.
type Terminator struct {
	quit chan struct{}
	done chan struct{}
}

func NewTerminator() *Terminator {
	return &Terminator{
		quit: make(chan struct{}),
		done: make(chan struct{}),
	}
}

// Quit returns a channel which acts a quit condition
// that can be waited upon by trying to receive from the channel
func (term *Terminator) Quit() <-chan struct{} {
	return term.quit
}

// Done signals complete termination of a task to the terminator
func (term *Terminator) Done() {
	close(term.done)
}

// Terminate signals a task to quit and waits for a completion signal
func (term *Terminator) Terminate() {
	close(term.quit)
	<-term.done
}
