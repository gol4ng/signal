package signal_test

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"bou.ke/monkey"

	"github.com/stretchr/testify/assert"

	sig "github.com/gol4ng/signal"
)

func TestKillerSubscribe_Kill(t *testing.T) {
	var killerSignalChan chan<- os.Signal
	var killerSubscribedSignals = []os.Signal{os.Kill, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGSTOP}
	var exitCalled = false

	monkey.Patch(os.Exit, func(code int) {
		assert.Equal(t, 1, code)
		exitCalled = true
	})
	monkey.Patch(signal.Notify, func(c chan<- os.Signal, signals ...os.Signal) {
		// get subscriber internal chan in order to simulate signal receive
		killerSignalChan = c
		for _, s := range killerSubscribedSignals {
			assert.Contains(t, signals, s)
		}
	})
	defer monkey.UnpatchAll()

	sig.KillerSubscriber()

	// simulate signal receive
	killerSignalChan <- os.Kill
	// wait for goroutine callback func fulfilment (@see subscriber.go `go callback(killerSubscribedSignals)`)
	time.Sleep(1 * time.Millisecond)
	assert.True(t, exitCalled, "exit func should be called.")
}

func TestKillerSubscribe_Interupt(t *testing.T) {
	var killerSignalChan chan<- os.Signal
	var killerSubscribedSignals = []os.Signal{os.Kill, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGSTOP}
	var exitCalled = false

	monkey.Patch(os.Exit, func(code int) {
		assert.Equal(t, 130, code)
		exitCalled = true
	})
	monkey.Patch(signal.Notify, func(c chan<- os.Signal, signals ...os.Signal) {
		// get subscriber internal chan in order to simulate signal receive
		killerSignalChan = c
		for _, s := range killerSubscribedSignals {
			assert.Contains(t, signals, s)
		}
	})
	defer monkey.UnpatchAll()

	sig.KillerSubscriber()

	// simulate signal receive
	killerSignalChan <- os.Interrupt
	// wait for goroutine callback func fulfilment (@see subscriber.go `go callback(killerSubscribedSignals)`)
	time.Sleep(1 * time.Millisecond)
	assert.False(t, exitCalled, "exit func should not be called when first interrupt raised.")
	killerSignalChan <- os.Interrupt
	time.Sleep(1 * time.Millisecond)
	assert.True(t, exitCalled, "exit func should be called.")
}

func TestSubscribeWithKiller_Interupt(t *testing.T) {
	var killerSignalChan chan<- os.Signal
	var killerSubscribedSignals = []os.Signal{os.Kill, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGSTOP}
	var exitCalled = false
	var signalChan chan<- os.Signal
	var subscribedSignal = os.Interrupt
	var callbackCalledTimes = 0
	var realSignal = os.Interrupt

	monkey.Patch(os.Exit, func(code int) {
		assert.Equal(t, 130, code)
		exitCalled = true
	})
	monkey.Patch(signal.Notify, func(c chan<- os.Signal, signals ...os.Signal) {
		if len(signals) == 1 {
			signalChan = c
			assert.Equal(t, subscribedSignal, signals[0])
			return
		}
		// get subscriber internal chan in order to simulate signal receive
		killerSignalChan = c
		for _, s := range killerSubscribedSignals {
			assert.Contains(t, signals, s)
		}
	})
	defer monkey.UnpatchAll()

	sig.SubscribeWithKiller(func(signal os.Signal) {
		assert.Equal(t, realSignal, signal, "wrong signal ingested by subscriber.")
		callbackCalledTimes++
	}, subscribedSignal)

	// simulate signal receive
	killerSignalChan <- realSignal
	signalChan <- realSignal
	// wait for goroutine callback func fulfilment (@see subscriber.go `go callback(killerSubscribedSignals)`)
	time.Sleep(1 * time.Millisecond)
	assert.False(t, exitCalled, "exit func should not be called yet.")
	assert.Equal(t, 1, callbackCalledTimes, "callback func should be called once.")
	killerSignalChan <- realSignal
	time.Sleep(1 * time.Millisecond)
	assert.True(t, exitCalled, "exit func should be called.")
	assert.Equal(t, 1, callbackCalledTimes, "callback func should be called only once.")
}

func TestUnKillerSubscribe(t *testing.T) {
	var killerSignalChan chan<- os.Signal
	var killerSubscribedSignals = []os.Signal{os.Kill, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGSTOP}
	var stopCalled = false
	var exitCalled = false

	monkey.Patch(os.Exit, func(code int) {
		t.Error("exit func must not be called.")
		exitCalled = true
	})
	monkey.Patch(signal.Notify, func(c chan<- os.Signal, signals ...os.Signal) {
		// get subscriber internal chan in order to simulate signal receive
		killerSignalChan = c
		for _, s := range killerSubscribedSignals {
			assert.Contains(t, signals, s)
		}
	})
	monkey.Patch(signal.Stop, func(c chan<- os.Signal) {
		assert.Equal(t, killerSignalChan, c)
		stopCalled = true
	})
	defer monkey.UnpatchAll()

	unsubscribeFunc := sig.KillerSubscriber()

	unsubscribeFunc()
	// wait for goroutine signal.Stop func call (@see subscriber.go `case <-stopChan:`)
	time.Sleep(1 * time.Millisecond)
	// `killerSignalChan` is a bidirectional chan as it is equal to the `subscriber killerSignalChan` value
	_, ok := <-*(*chan os.Signal)(unsafe.Pointer(&killerSignalChan))
	assert.False(t, ok, "signal channel should be closed.")
	assert.True(t, stopCalled, "`signal.Stop` should be called when `unsubscribeFunc` is called.")
	assert.False(t, exitCalled, "exit func should not be called as `unsubscribeFunc` is called.")
}
