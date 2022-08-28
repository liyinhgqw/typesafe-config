# Typesafe-conf

This project is a (very much incomplete) Golang version of typesafe config (HOCON syntax) for akka.

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
 - GetDuration
 - GetDefaultDuration

 
### Envs
 
Envs are parsed if they exist (syntax `... = ${?SOMEENV}` or `... = ${SOMEENV}`).

To use an env as an override:

    port = 9999
    port = ${?PORT}
        
In this case the variable will not be overridden if `PORT` is not defined.

In contrast

    port = 9999
    port = ${PORT}
    
In this case the variable will be overridden even if `PORT` isn't defined.

### Substitution

Substitutions (currently only values, not objects) are parsed if they exist (syntax `... = ${?some.path.to.value}` or `... = ${some.path.to.value}`).


    earlier-var = 100

    later-var = ${earlier-var}
    // results in `later-var = 100`
 
### Roadmap

 - Reusable sections
 
