package services

type ServiceErr struct {
	Err error
	Msg string
}

func (s *ServiceErr) Error() string {
	return s.Err.Error()
}

func (s *ServiceErr) Message() string {
	return s.Msg
}

func (s *ServiceErr) Unwrap() error {
	return s.Err
}
