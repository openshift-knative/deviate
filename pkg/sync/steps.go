package sync

type step func() error

type multiStep []step

func (m multiStep) runSteps() error {
	return runSteps(m)
}

func runSteps(steps []step) error {
	for _, st := range steps {
		if err := st(); err != nil {
			return err
		}
	}
	return nil
}
