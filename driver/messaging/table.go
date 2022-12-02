package messaging

type Table map[string]any

func (t Table) Validate() error {
	return nil
}
