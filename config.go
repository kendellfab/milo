package milo

// Configuration option to change the bind address.
func SetBind(bind string) func(*Milo) error {
	return func(m *Milo) error {
		m.bind = bind
		return nil
	}
}

// Configuration option to change the bind port.
func SetPort(port int) func(*Milo) error {
	return func(m *Milo) error {
		m.port = port
		return nil
	}
}

// Configuration option to change the port increment.
func SetPortInc(inc bool) func(*Milo) error {
	return func(m *Milo) error {
		m.portIncrement = inc
		return nil
	}
}
