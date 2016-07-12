package msgs

import "VODone/Tester/service"

func Send(ser service.Service, data []byte) (int, error) {
	lock := ser.GetLock()
	lock.Lock()

	// check write len(data) size buf
	n, err := ser.GetWriter().Write(data)
	if err != nil {
		lock.Unlock()
		return n, err
	}
	ser.GetWriter().Flush()

	lock.Unlock()
	return n, err
}
