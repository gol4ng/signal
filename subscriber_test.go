package signal_test

import (
	"os"
	"os/signal"
	"testing"
	"time"
	"unsafe"

	"bou.ke/monkey"

	"github.com/stretchr/testify/assert"

	sig "github.com/gol4ng/signal"
)

func TestSubscribe(t *testing.T) {
	var signalChan chan<- os.Signal
	var subscribedSignals = []os.Signal{os.Interrupt, os.Kill}
	var realSignal = os.Interrupt
	var callbackCalled = false

	monkey.Patch(signal.Notify, func(c chan<- os.Signal, signals ...os.Signal) {
		// get subscriber internal chan in order to simulate signal receive
		signalChan = c
		for _, s := range subscribedSignals {
			assert.Contains(t, signals, s)
		}
	})
	defer monkey.UnpatchAll()

	sig.Subscribe(func(signal os.Signal) {
		assert.Equal(t, realSignal, signal, "wrong signal ingested by subscriber.")
		callbackCalled = true
	}, subscribedSignals...)

	// simulate signal receive
	signalChan <- realSignal
	// wait for goroutine callback func fulfilment (@see subscriber.go `go callback(subscribedSignals)`)
	time.Sleep(1 * time.Millisecond)
	assert.True(t, callbackCalled, "callback func should be called.")
}

func TestUnSubscribe(t *testing.T) {
	var signalChan chan<- os.Signal
	var subscribedSignal = os.Interrupt
	var stopCalled = false
	var callbackCalled = false

	monkey.Patch(signal.Notify, func(c chan<- os.Signal, signals ...os.Signal) {
		// get subscriber internal chan in order to simulate signal receive
		signalChan = c
		assert.Equal(t, subscribedSignal, signals[0])
	})
	monkey.Patch(signal.Stop, func(c chan<- os.Signal) {
		assert.Equal(t, signalChan, c)
		stopCalled = true
	})
	defer monkey.UnpatchAll()

	unsubscribeFunc := sig.Subscribe(func(signal os.Signal) {
		assert.Equal(t, signal, subscribedSignal, "wrong signal ingested by subscriber.")
		callbackCalled = true
	}, subscribedSignal)

	unsubscribeFunc()
	// wait for goroutine signal.Stop func call (@see subscriber.go `case <-stopChan:`)
	time.Sleep(1 * time.Millisecond)
	// `signalChan` is a bidirectional chan as it is equal to the `subscriber signalChan` value
	_, ok := <-*(*chan os.Signal)(unsafe.Pointer(&signalChan))
	assert.False(t, ok, "signal channel should be closed.")
	assert.True(t, stopCalled, "`signal.Stop` should be called when `unsubscribeFunc` is called.")
	assert.False(t, callbackCalled, "callback func should not be called as `unsubscribeFunc` is called.")
}
