package reflection

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type (
	Ifc interface {
		GetName() string
		SetName(str string) string
	}

	IfcImpl struct {
		name string
	}

	Class1 struct {
		Field1 string
		field2 Class2
	}

	Class2 struct {
		field1 string
		field2 Ifc
		field3 *Class1
	}
)

func getInstance() *Class1 {
	instance := &Class1{
		Field1: "field1",
	}
	instance.field2 = Class2{
		field1: "field1",
		field2: &IfcImpl{name: "ifcImpl"},
		field3: instance,
	}
	return instance
}

func (i *IfcImpl) GetName() string {
	return i.name
}

func (i *IfcImpl) SetName(str string) string {
	old := i.name
	i.name = str
	return old
}

func (i *IfcImpl) setName(str string) string {
	return i.SetName(str)
}

func TestReflectedFields(t *testing.T) {
	instance := getInstance()
	{
		c1field1 := GetField[string](instance, "Field1")
		require.Equal(t, "field1", c1field1)
		c1field1Ptr := GetField[*string](instance, "Field1")
		*c1field1Ptr = "field1_changed"
		c1field1 = GetField[string](instance, "Field1")
		require.Equal(t, "field1_changed", c1field1)
	}

	{
		c1field2 := GetField[Class2](instance, "field2")
		require.Equal(t, "field1", c1field2.field1)
		c1field2.field1 = "field1_changed"
		c1field2 = GetField[Class2](instance, "field2")
		require.Equal(t, "field1", c1field2.field1)              // field2 is a (shallow) copy of the original field2
		require.Same(t, instance.field2.field2, c1field2.field2) // field2 is a reference to the original field2 (only the pointer was shallow-copied)
	}

	{
		c1field2Ptr := GetField[*Class2](instance, "field2")
		require.Equal(t, "field1", c1field2Ptr.field1)
		c1field2Ptr.field1 = "field1_changed"
		c1field2 := GetField[Class2](instance, "field2")
		require.Equal(t, "field1_changed", c1field2.field1)
	}

	{
		ifc := GetField[Ifc](instance, "field2", "field2")
		require.Equal(t, "ifcImpl", ifc.GetName())
		ifcImplPtr := ifc.(*IfcImpl)
		require.Equal(t, "ifcImpl", ifcImplPtr.name)
		ifcImpl := GetField[IfcImpl](instance, "field2", "field2")
		require.Equal(t, "ifcImpl", ifcImpl.name)
		ifcImplPtr = GetField[*IfcImpl](instance, "field2", "field2")
		require.Equal(t, "ifcImpl", ifcImplPtr.name)
		ifcImplPtr.name = "ifcImpl_changed"
		require.Equal(t, "ifcImpl_changed", GetField[Ifc](instance, "field2", "field2").GetName())
	}

	{
		c1field3 := GetField[*Class1](instance, "field2", "field3")
		require.Same(t, instance, c1field3)
		require.Same(t, instance.field2.field3, GetField[*Class1](instance, "field2", "field3", "field2", "field3"))
	}

}
