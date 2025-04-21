package store

type Object interface {
	ID() string
	Attributes() map[string]string
}

type object struct {
	id         string
	attributes map[string]string
}

func NewObject(id string, attributes map[string]string) Object {
	return &object{
		id,
		attributes,
	}
}

func (o *object) ID() string {
	return o.id
}

func (o *object) Attributes() map[string]string {
	return o.attributes
}
