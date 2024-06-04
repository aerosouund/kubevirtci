package cnao

type CnaoOpt struct{}

func NewCnaoOpt() *CnaoOpt {
	return &CnaoOpt{}
}

func (o *CnaoOpt) Exec() error {
	return nil
}
