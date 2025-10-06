package facade

type Master interface {
	ListSubjectAreas() ([]SubjectArea, error)
}

type SubjectArea struct {
	ID   int
	Name string
}
