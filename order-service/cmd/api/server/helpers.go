package server

func (r *MessageQueue) background(fn func()) {
	r.wg.Add(1)

	go func() {
		defer r.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				r.l.Error("%s", err)
			}
		}()

		fn()
	}()
}
