package storage

import (
	"testing"

	model "github.com/MonteCarloClub/dabe/model"
)

func TestSerialization(t *testing.T) {
	dabe := new(model.DABE)
	dabe.GlobalSetup()

	element_g1 := dabe.CurveParam.GetNewG1()
	serializer := NewElementSerializer()
	serializedElement, err := serializer.SerializeElement(element_g1)
	if err != nil {
		t.Errorf("Serialization failed: %v", err)
	}
	deserializedElement, err := serializer.DeserializeElement(serializedElement, dabe.CurveParam, "G1")
	if err != nil {
		t.Errorf("Deserialization failed: %v", err)
	}
	if element_g1.Equals(deserializedElement) == false {
		t.Errorf("Deserialized element does not match original")
	} else {
		t.Logf("Serialization and Deserialization successful for G1 element")
	}

	element_gt := dabe.CurveParam.GetNewGT()
	serializedElementGT, err := serializer.SerializeElement(element_gt)
	if err != nil {
		t.Errorf("Serialization failed: %v", err)
	}
	deserializedElementGT, err := serializer.DeserializeElement(serializedElementGT, dabe.CurveParam, "GT")
	if err != nil {
		t.Errorf("Deserialization failed: %v", err)
	}
	if element_gt.Equals(deserializedElementGT) == false {
		t.Errorf("Deserialized element does not match original")
	} else {
		t.Logf("Serialization and Deserialization successful for GT element")
	}

	element_zn := dabe.CurveParam.GetNewZn()
	serializedElementZn, err := serializer.SerializeElement(element_zn)
	if err != nil {
		t.Errorf("Serialization failed: %v", err)
	}
	deserializedElementZn, err := serializer.DeserializeElement(serializedElementZn, dabe.CurveParam, "Zn")
	if err != nil {
		t.Errorf("Deserialization failed: %v", err)
	}
	if element_zn.Equals(deserializedElementZn) == false {
		t.Errorf("Deserialized element does not match original")
	} else {
		t.Logf("Serialization and Deserialization successful for Zp element")
	}
}

func TestDabeSerialization(t *testing.T) {
	dabe := new(model.DABE)
	dabe.GlobalSetup()
	serializedDabe, err := SerializeDABE(dabe)
	if err != nil {
		t.Errorf("DABE Serialization failed: %v", err)
	}
	deserializedDabe, err := DeserializeDABE(serializedDabe)
	if err != nil {
		t.Errorf("DABE Deserialization failed: %v", err)
	}
	if dabe.G.Equals(deserializedDabe.G) == false || dabe.EGG.Equals(deserializedDabe.EGG) == false {
		t.Errorf("Deserialized DABE does not match original")
	} else {
		t.Logf("DABE Serialization and Deserialization successful")
	}

}
