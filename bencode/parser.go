package bencode

import (
	"bufio"
	"io"
)

// Bencode的反序列化
func Parse(r io.Reader) (*BObject, error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	//只查看第一位是什么并不读取(保护文件完整性)
	b, err := br.Peek(1)
	if err != nil {
		return nil, err
	}
	var ret BObject
	switch {
	//解析为string类型
	case b[0] >= '0' && b[0] <= '9':
		//parse string
		val, err := Decodestring(br)
		if err != nil {
			return nil, err
		}
		ret.type_ = BSTR
		ret.value_ = val
	//解析为int类型
	case b[0] == 'i':
		//parse int
		val, err := DecodeInt(br)
		if err != nil {
			return nil, err
		}
		ret.type_ = BINT
		ret.value_ = val
	case b[0] == 'l':
		//递归下降解析list
		br.ReadByte()
		var list []*BObject
		for {
			if p, _ := br.Peek(1); p[0] == 'e' {
				br.ReadByte()
				break
			}
			elem, err := Parse(br)
			if err != nil {
				return nil, err
			}
			list = append(list, elem)
		}
		ret.type_ = BLIST
		ret.value_ = list

	case b[0] == 'd':
		//递归下降解析dict
		br.ReadByte()
		dict := make(map[string]*BObject)
		for {
			if p, _ := br.Peek(1); p[0] == 'e' {
				br.ReadByte()
				break
			}
			key, err := Decodestring(br)
			if err != nil {
				return nil, err
			}
			val, err := Parse(br)
			if err != nil {
				return nil, err
			}
			dict[key] = val
		}
		ret.type_ = BDICT
		ret.value_ = dict
	default:
		return nil, ErrIvd
	}
	return &ret, nil
}
