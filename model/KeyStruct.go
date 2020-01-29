package MulticenterABEForFabric

import (
	"github.com/Nik-U/pbc"
)

/* Public Key Structure */
type PK struct {
	Gy       *pbc.Element //G^y, y from Zp
}

func (p *PK) Initialize(gy *pbc.Element) {
	p.Gy = gy
}

func (p *PK) getGy() *pbc.Element {
	return p.Gy.NewFieldElement().Set(p.Gy)
}

type SK struct {
	Y     *pbc.Element
}

func (s *SK) Initialize(y *pbc.Element) {
	s.Y = y
}

func (s *SK) getY() *pbc.Element {
	return s.Y.NewFieldElement().Set(s.Y)
}