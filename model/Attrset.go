package MulticenterABEForFabric

/* Attribute Set, Omega */
type AttrSet struct {
	Omega    []AttrSetNode
	Rhos_map map[string]uint32
	Size     uint32
}
type AttrSetNode struct {
	Name string
	Id uint32
}
func (this *AttrSet) Initialize(attrs_univ []string) () {
	this.Omega = make([]AttrSetNode,len(attrs_univ)+1,len(attrs_univ)+1)
	this.Rhos_map = make(map[string]uint32)
	for i:=0; i<len(this.Omega)-1 ; i++ {
		this.Omega[i+1] = NewAttrSetNode(attrs_univ[i], uint32(i+1))
		this.Rhos_map[attrs_univ[i]] = uint32(i+1)
	}

	this.Size = uint32(len(this.Omega))
}

func NewAttrSetNode(name string, id uint32) (AttrSetNode) {
	N := new(AttrSetNode)
	N.Name = name
	N.Id = id
	return *N
}
