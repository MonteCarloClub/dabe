package MulticenterABEForFabric

import "github.com/Nik-U/pbc"

type Cipher struct {
	C0         *pbc.Element
	C1s        []*pbc.Element
	C2s        []*pbc.Element
	C3s        []*pbc.Element
	CipherText []byte
	Policy     string
}
