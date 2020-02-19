package main

import (
	"encoding/json"
	"fmt"
	"github.com/thorweiyan/MulticenterABEForFabric/model"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

type MAFF int

var this *MulticenterABEForFabric.MAFFscheme

type state struct {
	SysInited  bool
	AASetup1ed bool
	AASetup2ed bool
	AASetup3ed bool
	Sij        map[string]bool
}

var states = state{false, false, false, false, make(map[string]bool)}

//server端函数如下，client端希望只看到下面的接口，即透明
type Sysinit struct {
	T, N int
}

func (maff *MAFF) SYSInit(args *Sysinit, reply *[]byte) error {
	fmt.Println("sysinit:")
	if !states.SysInited {
		states.SysInited = true
		fmt.Println("T: ", args.T, " N: ", args.N)
		this.SYSInit(args.T, args.N)
	} else {
		time.Sleep(5 * time.Second)
	}
	*reply = this.PublicKey.Serialize()
	return nil
}

type Setup1 struct {
	Pubkey []byte
	AAaid  []byte
}

func (maff *MAFF) AASetup1(args *Setup1, reply *Sysinit) error {
	fmt.Println("AASetup1:")
	if !states.AASetup1ed {
		states.AASetup1ed = true
		fmt.Println("aapubkey: ", args.AAaid)
		this.AASETUP1(args.Pubkey, args.AAaid)
	} else {
		time.Sleep(5 * time.Second)
	}
	reply.T = this.PublicKey.T
	reply.N = this.PublicKey.N
	return nil
}

func (maff *MAFF) AACommunicate(aaList []string, reply *[]string) error {
	fmt.Println("AACommunicate:")
	fmt.Printf("%x\n", aaList)
	fmt.Println(len(aaList))
	*reply = this.AACommunicate(aaList)
	fmt.Printf("%x\n", *reply)
	return nil
}

func (maff *MAFF) AppendSij(sij string, reply *bool) error {
	fmt.Println("AppendSij:")
	if _, ok := states.Sij[sij]; ok {
		*reply = false
		return nil
	} else {
		fmt.Printf("appendsij:%x\n", sij)
		this.AppendSij(sij)
		*reply = len(this.Sij) == this.PublicKey.N
		states.Sij[sij] = true
		return nil
	}
}

type Setup2 struct {
	Pki, Aid []byte
}

func (maff *MAFF) AASetup2(args string, reply *Setup2) error {
	fmt.Println("AASetup2:")
	if !states.AASetup2ed {
		states.AASetup2ed = true
		fmt.Println("aasetup2")
		this.AASETUP2()
	} else {
		time.Sleep(5 * time.Second)
	}
	reply.Pki = this.Pki.Bytes()
	reply.Aid = this.Aid.Bytes()
	fmt.Printf("pki:%x\n", this.Pki.Bytes())
	fmt.Printf("aid:%x\n", this.Aid.Bytes())
	return nil
}

type Setup3 struct {
	Pki, Aid [][]byte
}

func (maff *MAFF) AASetup3(args *Setup3, reply *[]byte) error {
	fmt.Println("AASetup3:")
	if !states.AASetup3ed {
		states.AASetup3ed = true
		fmt.Println("AASetup3:")
		fmt.Printf("%x\n", args.Pki)
		fmt.Printf("%x\n", args.Aid)
		this.AASETUP3(args.Pki, args.Aid)
		fmt.Printf("%x\n", this.PublicKey.E_gg_Alpha)
	} else {
		time.Sleep(5 * time.Second)
	}
	return nil
}

type Mmap struct {
	Map    []byte
	NowLen int
}

func (maff *MAFF) MarshalMap(args string, reply *Mmap) error {
	fmt.Println("MarshalMap:")
	attrs, err := json.Marshal(this.Omega.Rhos_map)
	if err != nil {
		*reply = Mmap{Map: nil, NowLen: 0}
		return fmt.Errorf("Marshal ABE's map error: " + err.Error())
	}
	*reply = Mmap{Map: attrs, NowLen: this.NowAttr}
	return nil
}

func (maff *MAFF) UnMarshalMap(args *Mmap, reply *[]byte) error {
	fmt.Println("UnMarshalMap:")
	temp := make(map[string]uint32)
	err := json.Unmarshal(args.Map, &temp)
	if err != nil {
		return fmt.Errorf("Unmarshal ABE's map error: " + err.Error())
	}
	this.ReplaceAttr(temp, args.NowLen)
	return nil
}

func (maff *MAFF) IsUserExists(userName string, reply *bool) error {
	fmt.Println("IsUserExists:", userName)
	if _, ok := this.Omega.Rhos_map[userName]; ok {
		*reply = true
		return nil
	}
	*reply = false
	return nil
}

func (maff *MAFF) AddAttr(userAttributes []string, reply *[]byte) error {
	fmt.Println("AddAttr:", userAttributes)
	return this.AddAttr(userAttributes)
}

func (maff *MAFF) PartUserSkGen(attrs []string, reply *[]byte) error {
	fmt.Println("PartUserSkGen:", attrs)
	for _, attr := range attrs {
		if _, ok := this.Omega.Rhos_map[attr]; !ok {
			fmt.Println(attr)
			fmt.Println(ok)
			*reply = nil
			return fmt.Errorf("don't have this attr!\n")
		}
	}
	re, err := this.SKGEN_AA(attrs)
	*reply = re
	return err
}

type Keygen struct {
	PartSk, Aid [][]byte
}

func (maff *MAFF) KeyGen(args *Keygen, reply *[]byte) error {
	fmt.Println("KeyGenByUser:")
	fmt.Println(args.Aid)
	*reply = this.SKGEN_USER(args.PartSk, args.Aid)
	return nil
}

type ENcrypt struct {
	Message, Policy string
}

func (maff *MAFF) Encrypt(args *ENcrypt, reply *[]byte) error {
	fmt.Println("Encrypt:")
	fmt.Println(args)
	*reply = this.ENCRYPT([]byte(args.Message), args.Policy)
	return nil
}

type DEcrypt struct {
	SecretKey, CryptText []byte
}

func (maff *MAFF) Decrypt(args *DEcrypt, reply *[]byte) error {
	fmt.Println("Decrypt:")
	re, err := this.DECRYPT(args.SecretKey, args.CryptText)
	*reply = re
	return err
}

func main() {
	this = new(MulticenterABEForFabric.MAFFscheme)

	maff := new(MAFF)
	rpc.Register(maff)
	rpc.HandleHTTP()

	var port string
	if len(os.Args) != 2 {
		fmt.Println("number of args ", len(os.Args), "is invalid!")
		return
	} else {
		port = os.Args[1]
		fmt.Println("listening in localhost:" + port)

		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

}
