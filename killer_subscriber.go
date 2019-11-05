package signal

import (
	"log"
	"os"
	"syscall"
)

// accordling to https://github.com/golang/go/issues/9463
// syscall.SIGSTOP, os.Kill are here for documentation purpose
var signalsKillers = []os.Signal{os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGSTOP, os.Kill}

// killer subscriber is global and unique
// it is unique because we only need one per application
var killerSubscriberFunc func()

// This is a simple signal subscriber that kills program
//
// It kills application when os.Kill or 2 interrupt signals are received
//
// Application gracefully stopped example:
// defer signal.SubscribeWithKiller(func(signal os.Signal) {
//   // here you can implement your application stopping steps
// }, os.Interrupt, syscall.SIGTERM)()
//
// Final user can now CTRL-C to stop your app gracefully
// and if CTRL-C is stroked a second time it is gonna kill the application
func KillerSubscriber() func() {
	if killerSubscriberFunc == nil {
		stopping := false
		killerSubscriberFunc = Subscribe(func(signal os.Signal) {
			if stopping {
				log.Println("killing application")
				os.Exit(130)
				return
			}
			println("Press `ctrl+c` again to kill application.")
			stopping = true
		}, signalsKillers...)
	}
	return killerSubscriberFunc
}

// This helper allows you to enable killer subscriber and subscribe your callback at once
func SubscribeWithKiller(callback func(os.Signal), signals ...os.Signal) func() {
	KillerSubscriber()
	return Subscribe(callback, signals...)
}
