package notify

import (
	"os/exec"

	"github.com/godbus/dbus/v5"
)

// Send delivers a desktop notification with the given summary and body.
// It tries the D-Bus freedesktop notifications interface first and falls
// back to notify-send if that fails.
func Send(summary, body string) error {
	if err := sendDbus(summary, body); err == nil {
		return nil
	}
	return sendNotifySend(summary, body)
}

func sendDbus(summary, body string) error {
	conn, err := dbus.SessionBusPrivate()
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.Auth(nil); err != nil {
		return err
	}
	if err := conn.Hello(); err != nil {
		return err
	}

	obj := conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")
	call := obj.Call(
		"org.freedesktop.Notifications.Notify", 0,
		"uberlauncher", uint32(0), "",
		summary, body,
		[]string{}, map[string]dbus.Variant{}, int32(-1),
	)
	return call.Err
}

func sendNotifySend(summary, body string) error {
	return exec.Command("notify-send", summary, body).Run()
}
