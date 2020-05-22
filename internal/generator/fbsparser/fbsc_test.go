package fbsparser

import (
	"github.com/objectbox/objectbox-go/internal/generator/fbsparser/reflection"
	"github.com/objectbox/objectbox-go/test/assert"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestFbsSchemaParser(t *testing.T) {
	schema, err := ParseSchemaFile("non-existent.fbs")
	assert.True(t, schema == nil)
	assert.Err(t, err)

	file, err := ioutil.TempFile("", "fbs-test")
	assert.NoErr(t, err)
	defer func() {
		assert.NoErr(t, os.Remove(file.Name()))
	}()

	_, err = file.WriteString(`
enum Planet:byte { Mercury = 0, Venus, Earth = 2 }

/// A real or imaginary living creature or entity
/// Note: name may be nil
table Being {
  age:short = 150;
  health:short = 100;
  name:string;
  friendly:bool = false (deprecated);
  location:Planet = Earth;

  /// All worldly belongings of this being
  belongings:[Item];
}

table Item {
  name:string;
  weight:short;
}

root_type Being;`)
	assert.NoErr(t, err)

	schema, err = ParseSchemaFile(file.Name())
	assert.NoErr(t, err)
	assert.True(t, schema != nil)

	assert.Eq(t, 1, schema.EnumsLength())
	assert.Eq(t, 2, schema.ObjectsLength())

	var enum reflection.Enum
	assert.True(t, schema.Enums(&enum, 0))
	assert.Eq(t, "Planet", string(enum.Name()))
	assert.Eq(t, 3, enum.ValuesLength())

	var enumVal reflection.EnumVal
	assert.True(t, enum.Values(&enumVal, 2))
	assert.Eq(t, "Earth", string(enumVal.Name()))

	var object reflection.Object
	assert.True(t, schema.Objects(&object, 1))
	assert.Eq(t, "Item", string(object.Name()))
	assert.Eq(t, 0, object.DocumentationLength())

	assert.True(t, schema.RootTable(&object) == &object)
	assert.Eq(t, "Being", string(object.Name()))

	assert.Eq(t, 2, object.DocumentationLength())
	assert.Eq(t, "A real or imaginary living creature or entity", strings.TrimSpace(string(object.Documentation(0))))
	assert.Eq(t, "Note: name may be nil", strings.TrimSpace(string(object.Documentation(1))))

	var field reflection.Field
	assert.Eq(t, 6, object.FieldsLength())
	assert.True(t, object.Fields(&field, 1)) // sorted by name
	assert.Eq(t, "belongings", string(field.Name()))

	assert.Eq(t, 1, field.DocumentationLength())
	assert.Eq(t, "All worldly belongings of this being", strings.TrimSpace(string(field.Documentation(0))))
}
