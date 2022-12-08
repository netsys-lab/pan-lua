package selector

import "github.com/netsec-ethz/scion-apps/pkg/pan"

type Selector interface {
	pan.Selector
	SetPreferences(map[string]string) error
}

type DefaultSelector struct {
	pan.DefaultSelector
}

func (s *DefaultSelector) SetPreferences(map[string]string) error {
	return nil
}

func (s *DefaultSelector) Path() *pan.Path {
	return s.DefaultSelector.Path()
}
