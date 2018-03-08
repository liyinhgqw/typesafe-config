package parse

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type Config struct {
	root Node
}

func (c *Config) GetValue(path string) (conf *Config, err error) {
	ps := strings.Split(path, ".")
	if len(ps) == 0 {
		err = errors.New("empty path")
		return
	} else {
		v := c.root
		for i := 0; i < len(ps); i++ {
			key := ps[i]
			if v.Type() == NodeMap {
				node, _ := v.(*MapNode)

				n, ok := node.Nodes[key]
				if !ok {
					err = errors.New("path not valid: " + key)
					return
				}

				for {
					if n.Type() == NodeField {
						fNode, ok := n.(*FieldNode)
						if !ok {
							err = errors.New("invalid field node: " + key)
							return
						}
						if cfg, nerr := c.GetValue(fNode.String()); nerr == nil {
							n = cfg.root
						} else if envV, ok := os.LookupEnv(fNode.String()); ok {
							n = &StringNode{Quoted: envV, NodeType: NodeString, Text: unquoteString(envV)}
						} else if fNode.Hard {
							n = &NilNode{NodeType: NodeNil}
						} else if fNode.Fallback != nil {
							n = fNode.Fallback
						} else {
							err = errors.New("invalid field node: " + key)
							return
						}
					} else {
						break
					}
				}
				v = n
			}
		}
		conf = &Config{root: v}
		return
	}
}

func (c *Config) String() string {
	return c.root.String()
}

func (c *Config) GetString(path string) (val string, err error) {
	conf, err := c.GetValue(path)
	if err != nil {
		return
	}
	if conf.root == nil {
		err = errors.New("not valid path: " + path)
	} else if conf.root.Type() == NodeString {
		if cstr, ok := conf.root.(*StringNode); ok {
			val = cstr.Text
		} else {
			err = errors.New("not valid string: " + cstr.String())
		}
	} else {
		err = errors.New("not valid string: " + path)
	}
	return
}

func (c *Config) GetDefaultString(path string, defaultVal string) string {
	val, err := c.GetString(path)
	if err != nil {
		return defaultVal
	}
	return val
}

func stripSpaces(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			// if the character is a space, drop it
			return -1
		}
		// else keep it in the string
		return r
	}, str)
}

func (c *Config) GetDuration(path string) (val time.Duration, err error) {
	str, err := c.GetString(path)
	if err != nil {
		return 0, err
	}

	str = stripSpaces(str)
	if len(str) == 0 {
		return
	}

	val, err = time.ParseDuration(str)
	return
}

func (c *Config) GetDefaultDuration(path string, defaultVal time.Duration) time.Duration {
	val, err := c.GetDuration(path)
	if err != nil {
		return defaultVal
	}
	return val
}

func (c *Config) GetBool(path string) (val bool, err error) {
	conf, err := c.GetValue(path)
	if err != nil {
		return
	}

	if conf.root.Type() == NodeBool {
		if cbool, ok := conf.root.(*BoolNode); ok {
			val = cbool.True
		} else {
			err = errors.New("not valid bool: " + cbool.String())
		}
	} else if conf.root.Type() == NodeString {
		if cstring, ok := conf.root.(*StringNode); ok {
			val, err = strconv.ParseBool(cstring.Text)
		} else {
			err = errors.New("not valid bool: " + cstring.String())
		}
	} else {
		err = errors.New("not valid bool: " + path)
	}
	return
}

func (c *Config) GetDefaultBool(path string, defaultVal bool) bool {
	val, err := c.GetBool(path)
	if err != nil {
		return defaultVal
	}
	return val
}

func (c *Config) GetInt(path string) (val int64, err error) {
	conf, err := c.GetValue(path)
	if err != nil {
		return
	}
	if conf.root.Type() == NodeNumber {
		if cnum, ok := conf.root.(*NumberNode); ok {
			switch {
			case cnum.IsInt:
				val = cnum.Int64
			default:
				err = errors.New("not valid int64: " + cnum.String())
			}
		} else {
			err = errors.New("not valid int64: " + cnum.String())
		}
	} else if conf.root.Type() == NodeString {
		if cstring, ok := conf.root.(*StringNode); ok {
			val, err = strconv.ParseInt(cstring.Text, 0, 64)
		} else {
			err = errors.New("not valid int64: " + cstring.String())
		}
	} else {
		err = errors.New("not valid int64: " + path)
	}
	return
}

func (c *Config) GetDefaultInt(path string, defaultVal int64) int64 {
	val, err := c.GetInt(path)
	if err != nil {
		return defaultVal
	}
	return val
}

func (c *Config) GetUInt(path string) (val uint64, err error) {
	conf, err := c.GetValue(path)
	if err != nil {
		return
	}
	if conf.root.Type() == NodeNumber {
		if cnum, ok := conf.root.(*NumberNode); ok {
			switch {
			case cnum.IsUint:
				val = cnum.Uint64
			default:
				err = errors.New("not valid uint64: " + cnum.String())
			}
		} else {
			err = errors.New("not valid uint64: " + cnum.String())
		}
	} else if conf.root.Type() == NodeString {
		if cstring, ok := conf.root.(*StringNode); ok {
			val, err = strconv.ParseUint(cstring.Text, 0, 64)
		} else {
			err = errors.New("not valid uint64: " + cstring.String())
		}
	} else {
		err = errors.New("not valid uint64: " + path)
	}
	return
}

func (c *Config) GetDefaultUInt(path string, defaultVal uint64) uint64 {
	val, err := c.GetUInt(path)
	if err != nil {
		return defaultVal
	}
	return val
}

func (c *Config) GetFloat(path string) (val float64, err error) {
	conf, err := c.GetValue(path)
	if err != nil {
		return
	}
	if conf.root.Type() == NodeNumber {
		if cnum, ok := conf.root.(*NumberNode); ok {
			switch {
			case cnum.IsFloat:
				val = cnum.Float64
			default:
				err = errors.New("not valid float64: " + cnum.String())
			}
		} else {
			err = errors.New("not valid float64: " + cnum.String())
		}
	} else if conf.root.Type() == NodeString {
		if cstring, ok := conf.root.(*StringNode); ok {
			val, err = strconv.ParseFloat(cstring.Text, 64)
		} else {
			err = errors.New("not valid float64: " + cstring.String())
		}
	} else {
		err = errors.New("not valid float64: " + path)
	}
	return
}

func (c *Config) GetDefaultFloat(path string, defaultVal float64) float64 {
	val, err := c.GetFloat(path)
	if err != nil {
		return defaultVal
	}
	return val
}

func (c *Config) GetComplex(path string) (val complex128, err error) {
	conf, err := c.GetValue(path)
	if err != nil {
		return
	}
	if conf.root.Type() == NodeNumber {
		if cnum, ok := conf.root.(*NumberNode); ok {
			switch {
			case cnum.IsComplex:
				val = cnum.Complex128
			default:
				err = errors.New("not valid complex: " + cnum.String())
			}
		} else if conf.root.Type() == NodeString {
			if cstring, ok := conf.root.(*StringNode); ok {
				if _, err := fmt.Sscan(cstring.Text, &val); err != nil {
					err = errors.New("not valid complex: " + cstring.String())
				}
			} else {
				err = errors.New("not valid complex: " + cstring.String())
			}
		} else {
			err = errors.New("not valid complex: " + cnum.String())
		}
	} else {
		err = errors.New("not valid complex: " + path)
	}
	return
}

func (c *Config) GetDefaultComplex(path string, defaultVal complex128) complex128 {
	val, err := c.GetComplex(path)
	if err != nil {
		return defaultVal
	}
	return val
}

func (c *Config) GetArray(path string) (vals []*Config, err error) {
	conf, err := c.GetValue(path)
	if err != nil {
		return
	}
	if conf.root.Type() == NodeList {
		if clist, ok := conf.root.(*ListNode); ok {
			for ind, n := range clist.Nodes {
				for {
					if n.Type() == NodeField {
						fNode, ok := n.(*FieldNode)
						if !ok {
							err = errors.New(fmt.Sprintf("invalid list node: %s[%d]", path, ind))
							return
						}
						if cfg, nerr := c.GetValue(fNode.String()); nerr == nil {
							n = cfg.root
						} else if envV, ok := os.LookupEnv(fNode.String()); ok {
							n = &StringNode{Quoted: envV, NodeType: NodeString, Text: unquoteString(envV)}
						} else if fNode.Hard {
							n = &NilNode{NodeType: NodeNil}
						} else if fNode.Fallback != nil {
							n = fNode.Fallback
						} else {
							err = errors.New(fmt.Sprintf("invalid field node: %s[%d]", path, ind))
							return
						}

					} else {
						break
					}
				}
				vals = append(vals, &Config{root: n})
			}
		} else {
			err = errors.New("not valid list node: " + clist.String())
		}
	} else {
		err = errors.New("not valid list node: " + path)
	}
	return
}
