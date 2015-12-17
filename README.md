# Typesafe-conf

This project is a Golang version of typesafe config for akka.

Can also populate a struct from a config file.

There are 2 ways to use this library, one is via tagging a struct and populating it, the other is manually accessing keys.

### Using populate

Tries to set fields on a struct using values from a config object.

 - struct names are dasherized when looking up config name
 - an initial prefix tells the function where to start looking from
 - tags can be used to give an alternative config field name, eg `config:"field-name" would be looked up in config as 'field-name'
 - tags can also be used to give a default, eg `config:"field-name,10" would set a default of 10
 
        import "github.com/byrnedo/typesafe-config/parse"
 		
        var testStruct = struct {
            Foo int
            Bar float64 `config:"barr"`
            SubStruct struct {
                Baz uint `config:"baz,10"`
            }
        }
 		
    	if tree, err := parse.ParseFile("./test.conf"); err != nil {
    		t.Error("Failed to read ./test.conf:" + err.Error())
    	}
    
    	parse.Populate(testStruct, tree.GetConfig(), "root")

### Manually

The following methods exist for manually accessing config values.

 - GetValue
 - GetString
 - GetDefaultString
 - GetBool
 - GetDefaultBool
 - GetInt
 - GetDefaultInt
 - GetUInt
 - GetDefaultUInt
 - GetFloat
 - GetDefaultFloat
 - GetComplex
 - GetDefaultComplex
 - GetArray
 
### Roadmap

 - Env parsing and with fallback.
 - Reusable sections
 