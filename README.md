# signal
This repository provides helpers with signals

## Subscriber
Signal subscriber that allows you to attach a callback to an `os.Signal` notification.

Useful to react to any os.Signal.

It returns an `unsubscribe` function that can gracefully stop some http server and clean allocated object
