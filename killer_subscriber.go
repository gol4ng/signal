package signal

import (
	"log"
	"os"
	"syscall"
)

var signalsKillers = []os.Signal{os.Kill, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGSTOP}

// killer subscriber is global and unique
// it is unique because we only need once per application
var killerSubscriber func()

// This is a simple signal subscriber that kill program
//
// It kill application when os.Kill or 2 interupt signal received
//
// Application gracefully stopped example:
// defer signal.SubscribeWithKiller(func(signal os.Signal) {
//   // delete tempoary file
//   // clean memory/cache etc
//   fmt.Println("implement your application stopping steps")
// }, os.Interrupt, syscall.SIGTERM)()
//
// Final user can now CTRL-C to stoppped gracefully your app
// and if he CTRL-C second time it gonna to kill application
func KillerSubscriber() func() {
	if killerSubscriber == nil {
		stopping := false
		killerSubscriber = Subscribe(func(signal os.Signal) {
			if signal == os.Kill {
				log.Println("killing application")
				os.Exit(1)
				return
			}
			if stopping {
				log.Println("killing application")
				os.Exit(130)
				return
			}
			println("Press `ctrl+c` again to kill application.")
			stopping = true
		}, signalsKillers...)
	}
	return killerSubscriber
}

// This helper allow you to enable killer subscriber and subscribe your callback at once
func SubscribeWithKiller(callback func(os.Signal), signals ...os.Signal) func() {
	KillerSubscriber()
	return Subscribe(callback, signals...)
}
