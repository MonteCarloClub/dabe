package DecentralizedABE

import (
	"fmt"
	"testing"
)

/*func TestReflect(t *testing.T) {
	helper = new(Helper)
	dabe := new(DABE)
	dabe.GlobalSetup()
	fudanUniversity := dabe.UserSetup("Fudan_University")
	fmt.Printf("%v\n", fudanUniversity)
	s := helper.Struct2Map(fudanUniversity)
	bytes, _ := json.Marshal(s)
	fmt.Println(bytes)
	user := new(User)
	helper.Str2Struct(bytes, user)
	fmt.Printf("%v\n", user)
	assert.Equal(t, fudanUniversity, user, "reflect error")
}*/

func TestReflect2(t *testing.T) {
	dabe := new(DABE)
	dabe.GlobalSetup()
	fudanUniversity := dabe.UserSetup("Fudan_University")
	fudanUniversity.GenerateNewAttr("Fudan_University:在读研究生", dabe)
	fmt.Printf("%v\n", fudanUniversity)
	bytes, err := Serialize2Bytes(fudanUniversity)
	if err != nil {
		panic(err)
	}
	fmt.Println(bytes)
	user := new(User)
	err = Deserialize2Struct(bytes, user)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
