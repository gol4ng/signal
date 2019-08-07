package signal

import (
	"log"
	"os"
	"syscall"
)

var signalsKiller = []os.Signal{os.Kill, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGSTOP}
var killerSubscriber func()

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
		}, signalsKiller...)
	}
	return killerSubscriber
}

func SubscribeWithKiller(callback func(os.Signal), signals ...os.Signal) func() {
	KillerSubscriber()
	return Subscribe(callback, signals...)
}
