package notify

type Listener interface {
	Listen(chan<- Message) error
	Close() error
}
