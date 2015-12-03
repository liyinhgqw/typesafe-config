package parse
import "strings"
import "errors"

type Config struct {
    root Node
}

func (c *Config) GetValue(path string) (conf *Config, err error) {
    ps := strings.Split(path, ".")
    if (len(ps) == 0) {
        err = errors.New("empty path")
        return
    } else {
        v := c.root
        for i := 0; i < len(ps); i++ {
            key := ps[i]
            if (v.Type() == NodeMap) {
                node, _ := v.(*MapNode)
                if n, ok := node.Nodes[key]; !ok {
                    err = errors.New("path not valid: " + key)
                    return
                } else {
                    v = n
                }
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
    if conf.root.Type() == NodeString {
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
            for _, n := range clist.Nodes {
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

