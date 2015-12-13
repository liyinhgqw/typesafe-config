package parse
import (
"reflect"
"strings"
	"strconv"
	"bytes"
	"fmt"
	"os"
	"unicode/utf8"
	"unicode"
)
// Tries to set fields on a struct using values from a config object.
//
// - struct names are dasherized when looking up config name
// - an initial prefix tells the function where to start looking from
// - tags can be used to give an alternative config field name ,
// 		eg `config:"field-name" would be looked up in config as 'field-name'
// - tags can also be used to give a default
// 		eg `config:"field-name,10" would set a default of 10
//
// Example:
//
// var x := struct {
// 		Foo int
//		Bar float64 `config:"barr"`
//		SubStruct struct {
//			Baz uint `config:"baz,10"`
// 		}
// }
// config.Populate(&x, "root")
//
// would be populated fully from the following config file:
//
// root {
//		foo = 9
//		barr = 0.9
//		sub-struct {
// 		}
//
// }
func Populate(targetPtr interface{}, conf *Config, prefix string) {
	setValue(reflect.ValueOf(targetPtr), conf, prefix, "", false)
}

func configFieldNamer(field reflect.StructField, prefix string) (name string, defaultVal string, hasDefault bool) {
	t := field.Tag.Get("config")
	tArr := strings.Split(t, ",")

	name = toDashCase(field.Name)

	if len(tArr) > 0 && len(tArr[0]) > 0 {
		switch tArr[0] {
		case "-":
			return "", "", false
		default:
			name = tArr[0]
		}
	}

	if len(tArr) > 1 {
		defaultVal = tArr[1]
		hasDefault = true
	}

	if len(prefix) > 0 {
		prefix = prefix + "."
	}
	return prefix + name, defaultVal, hasDefault
}

func setValue(field reflect.Value, conf *Config, configName string, defaultVal string, hasDefault bool) {
	var err error

	if field.Kind() != reflect.Ptr {
		panic("Not a pointer value " + field.Kind().String())
	}

	field = reflect.Indirect(field)
	if !field.CanSet() {
		return
	}
	switch field.Kind() {
	case reflect.Struct:
		itemType := reflect.TypeOf(field.Interface())

		for i := 0; i < field.NumField(); i++ {
			configFieldName, defaultVal, hasDefault := configFieldNamer(itemType.Field(i), configName)

			setValue(field.Field(i).Addr(), conf, configFieldName, defaultVal, hasDefault)
		}
	case reflect.Bool:
		var boolVal bool
		if hasDefault {
			defaultBool, _ := strconv.ParseBool(defaultVal)
			boolVal = conf.GetDefaultBool(configName, defaultBool)
		} else {
			boolVal, err = conf.GetBool(configName)
		}
		if err == nil {
			field.SetBool(boolVal)
		}
	case reflect.String:
		var strVal string
		if hasDefault {
			strVal = conf.GetDefaultString(configName, defaultVal)
		} else {
			strVal, err = conf.GetString(configName)
		}
		if err == nil {
			field.SetString(strVal)
		}
	case reflect.Int:
		var intVal int64
		if hasDefault {
			defaultInt, _ := strconv.Atoi(defaultVal)
			intVal = conf.GetDefaultInt(configName, int64(defaultInt))
		} else {
			intVal, err = conf.GetInt(configName)
		}
		if err == nil {
			field.SetInt(intVal)
		}
	case reflect.Int8:
		err = setIntVal(&field, conf, 8, configName, defaultVal, hasDefault)
	case reflect.Int16:
		err = setIntVal(&field, conf, 16, configName, defaultVal, hasDefault)
	case reflect.Int32:
		err = setIntVal(&field, conf, 32, configName, defaultVal, hasDefault)
	case reflect.Int64:
		err = setIntVal(&field, conf, 64, configName, defaultVal, hasDefault)
	case reflect.Uint:
		var uintVal uint64
		if hasDefault {
			defaultInt, _ := strconv.Atoi(defaultVal)
			uintVal = conf.GetDefaultUInt(configName, uint64(defaultInt))
		} else {
			uintVal, err = conf.GetUInt(configName)

		}
		if err == nil {
			field.SetUint(uintVal)
		}
	case reflect.Uint8:
		err = setUintVal(&field, conf, 8, configName, defaultVal, hasDefault)
	case reflect.Uint16:
		err = setUintVal(&field, conf, 16, configName, defaultVal, hasDefault)
	case reflect.Uint32:
		err = setUintVal(&field, conf, 32, configName, defaultVal, hasDefault)
	case reflect.Uint64:
		err = setUintVal(&field, conf, 64, configName, defaultVal, hasDefault)
	case reflect.Float32, reflect.Float64:
		var floatVal float64
		if hasDefault {
			var bits int
			if field.Kind() == reflect.Float32 {
				bits = 32
			} else {
				bits = 64
			}
			defaultFloat, _ := strconv.ParseFloat(defaultVal, bits)
			floatVal = conf.GetDefaultFloat(configName, defaultFloat)
		} else {
			floatVal, err = conf.GetFloat(configName)
		}
		if err == nil {
			field.SetFloat(floatVal)
		}
	case reflect.Slice:
		setSliceVal(&field, conf, configName)
	default:
	}

	if err != nil && ! strings.HasPrefix(err.Error(), "path not valid:"){
		fmt.Fprintf(os.Stderr, "Error reading config from path %s: %s\n", configName, err)
	}

}

func setSliceVal(field *reflect.Value, conf *Config, configName string) {
	confArr, err := conf.GetArray(configName)
	if err != nil {
		return
	}

	var newSlice reflect.Value
	switch field.Type().Elem().Kind() {
	case reflect.String:
		newSlice = reflect.MakeSlice(reflect.TypeOf([]string{}), 0, 0)
	case reflect.Float32:
		newSlice = reflect.MakeSlice(reflect.TypeOf([]float32{}), 0, 0)
	case reflect.Float64:
		newSlice = reflect.MakeSlice(reflect.TypeOf([]float64{}), 0, 0)
	case reflect.Int:
		newSlice = reflect.MakeSlice(reflect.TypeOf([]int{}), 0, 0)
	case reflect.Int8:
		newSlice = reflect.MakeSlice(reflect.TypeOf([]int8{}), 0, 0)
	case reflect.Int16:
		newSlice = reflect.MakeSlice(reflect.TypeOf([]int16{}), 0, 0)
	case reflect.Int32:
		newSlice = reflect.MakeSlice(reflect.TypeOf([]int32{}), 0, 0)
	case reflect.Int64:
		newSlice = reflect.MakeSlice(reflect.TypeOf([]int64{}), 0, 0)
	case reflect.Uint:
		newSlice = reflect.MakeSlice(reflect.TypeOf([]uint{}), 0, 0)
	case reflect.Uint8:
		newSlice = reflect.MakeSlice(reflect.TypeOf([]uint8{}), 0, 0)
	case reflect.Uint16:
		newSlice = reflect.MakeSlice(reflect.TypeOf([]uint16{}), 0, 0)
	case reflect.Uint32:
		newSlice = reflect.MakeSlice(reflect.TypeOf([]uint32{}), 0, 0)
	case reflect.Uint64:
		newSlice = reflect.MakeSlice(reflect.TypeOf([]uint64{}), 0, 0)
	}
	for i, item := range confArr {
		switch field.Type().Elem().Kind() {
		case reflect.Float32:
			val, err := item.GetFloat("")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read slice index [%d] for config %s: %s\n", i, configName, err)
			} else {
				newSlice = reflect.Append(newSlice, reflect.ValueOf(float32(val)))
			}
		case reflect.Float64:
			val, err := item.GetFloat("")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read slice index [%d] for config %s: %s\n", i, configName, err)
			} else {
				newSlice = reflect.Append(newSlice, reflect.ValueOf(float64(val)))
			}
		case reflect.String:
			val, err := item.GetString("")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read slice index [%d] for config %s: %s\n", i, configName, err)
			} else {
				newSlice = reflect.Append(newSlice, reflect.ValueOf(val))
			}
		case reflect.Int:
			val, err := item.GetInt("")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read slice index [%d] for config %s: %s\n", i, configName, err)
			} else {
				newSlice = reflect.Append(newSlice, reflect.ValueOf(int(val)))
			}
		case reflect.Int8:
			val, err := item.GetInt("")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read slice index [%d] for config %s: %s\n", i, configName, err)
			} else {
				newSlice = reflect.Append(newSlice, reflect.ValueOf(int8(val)))
			}
		case reflect.Int16:
			val, err := item.GetInt("")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read slice index [%d] for config %s: %s\n", i, configName, err)
			} else {
				newSlice = reflect.Append(newSlice, reflect.ValueOf(int16(val)))
			}
		case reflect.Int32:
			val, err := item.GetInt("")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read slice index [%d] for config %s: %s\n", i, configName, err)
			} else {
				newSlice = reflect.Append(newSlice, reflect.ValueOf(int32(val)))
			}
		case reflect.Int64:
			val, err := item.GetInt("")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read slice index [%d] for config %s: %s\n", i, configName, err)
			} else {
				newSlice = reflect.Append(newSlice, reflect.ValueOf(int64(val)))
			}
		case reflect.Uint:
			val, err := item.GetUInt("")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read slice index [%d] for config %s: %s\n", i, configName, err)
			} else {
				newSlice = reflect.Append(newSlice, reflect.ValueOf(uint(val)))
			}
		case reflect.Uint8:
			val, err := item.GetUInt("")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read slice index [%d] for config %s: %s\n", i, configName, err)
			} else {
				newSlice = reflect.Append(newSlice, reflect.ValueOf(uint8(val)))
			}
		case reflect.Uint16:
			val, err := item.GetUInt("")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read slice index [%d] for config %s: %s\n", i, configName, err)
			} else {
				newSlice = reflect.Append(newSlice, reflect.ValueOf(uint16(val)))
			}
		case reflect.Uint32:
			val, err := item.GetUInt("")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read slice index [%d] for config %s: %s\n", i, configName, err)
			} else {
				newSlice = reflect.Append(newSlice, reflect.ValueOf(uint32(val)))
			}
		case reflect.Uint64:
			val, err := item.GetUInt("")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read slice index [%d] for config %s: %s\n", i, configName, err)
			} else {
				newSlice = reflect.Append(newSlice, reflect.ValueOf(uint64(val)))
			}
		default:
			fmt.Println(field.Type().Elem().Kind())

		}
	}
	if field.CanSet() {
		field.Set(newSlice)
	}

}

func setIntVal(field *reflect.Value, conf *Config, bits int, configName string, defaultVal string, hasDefault bool) (err error) {
	var (
		intVal int64
	)
	if hasDefault {
		defaultInt, _ := strconv.ParseInt(defaultVal, 10, bits)
		intVal = conf.GetDefaultInt(configName, defaultInt)
	} else {
		intVal, err = conf.GetInt(configName)
	}
	if err == nil {
		field.SetInt(intVal)
	}
	return
}

func setUintVal(field *reflect.Value, conf *Config, bits int, configName string, defaultVal string, hasDefault bool) (err  error ){
	var (
		intVal uint64
	)
	if hasDefault {
		defaultInt, _ := strconv.ParseUint(defaultVal, 10, bits)
		intVal = conf.GetDefaultUInt(configName, defaultInt)
	} else {
		intVal, err = conf.GetUInt(configName)
	}
	if err == nil {
		field.SetUint(intVal)
	}
	return
}

//
// Copyright (c) 2015 Huan Du
// Modifications: 2015 Donal Byrne
func toDashCase(str string) string {
	if len(str) == 0 {
		return ""
	}

	buf := &bytes.Buffer{}
	var prev, r0, r1 rune
	var size int

	r0 = '-'

	for len(str) > 0 {
		prev = r0
		r0, size = utf8.DecodeRuneInString(str)
		str = str[size:]

		switch {
		case r0 == utf8.RuneError:
			buf.WriteByte(byte(str[0]))

		case unicode.IsUpper(r0):
			if prev != '-' {
				buf.WriteRune('-')
			}

			buf.WriteRune(unicode.ToLower(r0))

			if len(str) == 0 {
				break
			}

			r0, size = utf8.DecodeRuneInString(str)
			str = str[size:]

			if !unicode.IsUpper(r0) {
				buf.WriteRune(r0)
				break
			}

			// find next non-upper-case character and insert `_` properly.
			// it's designed to convert `HTTPServer` to `http_server`.
			// if there are more than 2 adjacent upper case characters in a word,
			// treat them as an abbreviation plus a normal word.
			for len(str) > 0 {
				r1 = r0
				r0, size = utf8.DecodeRuneInString(str)
				str = str[size:]

				if r0 == utf8.RuneError {
					buf.WriteRune(unicode.ToLower(r1))
					buf.WriteByte(byte(str[0]))
					break
				}

				if !unicode.IsUpper(r0) {
					if r0 == '_' || r0 == ' ' || r0 == '-' {
						r0 = '-'

						buf.WriteRune(unicode.ToLower(r1))
					} else {
						buf.WriteRune('-')
						buf.WriteRune(unicode.ToLower(r1))
						buf.WriteRune(r0)
					}

					break
				}

				buf.WriteRune(unicode.ToLower(r1))
			}

			if len(str) == 0 || r0 == '-' {
				buf.WriteRune(unicode.ToLower(r0))
				break
			}

		default:
			if r0 == ' ' || r0 == '_' {
				r0 = '-'
			}

			buf.WriteRune(r0)
		}
	}

	return buf.String()
}
