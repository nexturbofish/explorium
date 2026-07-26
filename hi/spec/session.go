package spec

type SessionStore interface {
	Save(session *Session) error
	Load(id string) (*Session, error)
	List() ([]SessionMeta, error)
	Delete(id string) error
}
