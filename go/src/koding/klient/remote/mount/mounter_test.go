package mount

import (
	"errors"
	"syscall"
	"testing"
	"time"

	"koding/klient/remote/kitepinger"
	"koding/klient/remote/machine"
	"koding/klient/remote/req"
	"koding/klient/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHandleChangeSummary(t *testing.T) {
	Convey("Given a Mounter", t, func() {
		fakeMount := &Mount{
			MountFolder: req.MountFolder{
				LocalPath: "fakeDir",
			},
		}
		mounter := Mounter{
			Log:     testutil.DiscardLogger,
			Machine: &machine.Machine{},
		}

		Convey("Given an OldStatus of Failure from 35 minutes ago", func() {
			changeSum := kitepinger.ChangeSummary{
				OldStatus:    kitepinger.Failure,
				OldStatusDur: 35 * time.Minute,
			}

			Convey("It should set the status to remounting while remounting", func() {
				// Because the value is set, and then unset, this logic is a bit weird.
				// We're blocking the call so that we can check the value in another thread while it is blocked.
				// We don't really care when it's done, the other tests cover that.
				canCheck := make(chan bool)
				go mounter.handleChangeSummary(fakeMount, changeSum, func(*Mount) error {
					canCheck <- true
					time.Sleep(1 * time.Second)
					return nil
				})
				<-canCheck
				close(canCheck)

				status, msg := mounter.Machine.GetStatus()
				So(status, ShouldEqual, machine.MachineRemounting)
				So(msg, ShouldEqual, autoRemounting)
			})

			Convey("When the remountFunc fails", func() {
				remountFunc := func(*Mount) error { return errors.New("err") }

				Convey("It should set the machine status to error", func() {
					err := mounter.handleChangeSummary(fakeMount, changeSum, remountFunc)
					So(err, ShouldNotBeNil)
					status, msg := mounter.Machine.GetStatus()
					So(status, ShouldEqual, machine.MachineError)
					So(msg, ShouldEqual, autoRemountFailed)
				})
			})

			Convey("When the remountFunc succeeds", func() {
				remountFunc := func(*Mount) error { return nil }

				Convey("It should clear the machine status", func() {
					err := mounter.handleChangeSummary(fakeMount, changeSum, remountFunc)
					So(err, ShouldBeNil)
					status, msg := mounter.Machine.GetStatus()
					// We can't know the real value here, because GetStatus looks up
					// online/offline, so instead we'll just make sure we're not still
					// remounting status.
					So(status, ShouldNotEqual, machine.MachineRemounting)
					So(msg, ShouldEqual, "")
				})
			})
		})

		Convey("Given an OldStatus of Failure less than 30 minutes ago", func() {
			remountFunc := func(*Mount) error { return nil }
			changeSum := kitepinger.ChangeSummary{
				OldStatus:    kitepinger.Failure,
				OldStatusDur: 28 * time.Minute,
			}

			// It's hard to know if the status was changed, so lets set an unlikely status
			// to serve as a test.
			mounter.Machine.SetStatus(machine.MachineConnected, "foobarbaz")
			Convey("It should not change the remote status", func() {
				err := mounter.handleChangeSummary(fakeMount, changeSum, remountFunc)
				So(err, ShouldBeNil)
				status, msg := mounter.Machine.GetStatus()
				So(status, ShouldEqual, machine.MachineConnected)
				So(msg, ShouldEqual, "foobarbaz")
			})

			Convey("It should not call the remounter", func() {
				remountFunc := func(*Mount) error {
					return errors.New("remount should not be called")
				}
				err := mounter.handleChangeSummary(fakeMount, changeSum, remountFunc)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestRetryOnConnErr(t *testing.T) {
	Convey("Given a function that returns a conn error", t, func() {
		retryCount := 0
		retry := func() error {
			retryCount++
			return syscall.ECONNREFUSED
		}
		retryUntil3 := func() error {
			retryCount++
			if retryCount >= 3 {
				return nil
			}
			return syscall.ECONNREFUSED
		}

		Convey("It should retry the specified number of times", func() {
			retryOnConnErr(3, 1, retry)
			So(retryCount, ShouldEqual, 3)
		})

		Convey("It should return the error after max attempts", func() {
			So(retryOnConnErr(3, 1, retry), ShouldEqual, syscall.ECONNREFUSED)
		})

		Convey("It should stop once the func succeeds", func() {
			retryOnConnErr(4, 1, retryUntil3)
			So(retryCount, ShouldEqual, 3)
		})

		Convey("It should not return an error if the func eventually succeeds", func() {
			So(retryOnConnErr(4, 1, retryUntil3), ShouldBeNil)
		})
	})

	Convey("Given a function that does not return a conn error", t, func() {
		retryCount := 0
		retry := func() error {
			retryCount++
			return nil
		}

		Convey("It should try once", func() {
			retryOnConnErr(3, 1, retry)
			So(retryCount, ShouldEqual, 1)
		})

		Convey("It should not return an error", func() {
			So(retryOnConnErr(3, 1, retry), ShouldBeNil)
		})
	})
}