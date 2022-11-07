package bencode

import (
	"bufio"
	"errors"
	"io"
)

var (
	ErrNum = errors.New("expect num")
	ErrCol = errors.New("expect colon")
	ErrEpI = errors.New("expect char i")
	ErrEpE = errors.New("expect char e")
	ErrTyp = errors.New("wrong type")
	ErrIvd = errors.New("invalid bencode")
)

// bencode只有四种数据类型 所以uint8存储就可以
type BType uint8

// 四种数据类型的状态码
const (
	BSTR  BType = 0x01
	BINT  BType = 0x02
	BLIST BType = 0x03
	BDICT BType = 0x04
)

// 四种Btype所对应的值 可以使用泛型但并未使用
type BValue interface{}

// 整体的bencode解析对应的结构
type BObject struct {
	type_  BType
	value_ BValue
}

// Str 对应四种数据类型的方法
// 方法调用时 因为要修改对应的值,所以传入指针
func (o *BObject) Str() (string, error) {
	if o.type_ != BSTR {
		return "", errors.New("wrong type")
	}
	return o.value_.(string), nil
}

func (o *BObject) Int() (int, error) {
	if o.type_ != BINT {
		return 0, errors.New("wrong type")
	}
	return o.value_.(int), nil
}

func (o *BObject) List() ([]*BObject, error) {
	if o.type_ != BLIST {
		return nil, errors.New("wrong type")
	}
	return o.value_.([]*BObject), nil
}

func (o *BObject) Dict() (map[string]*BObject, error) {
	if o.type_ != BDICT {
		return nil, errors.New("wrong type")
	}
	return o.value_.(map[string]*BObject), nil
}

// Bencode的序列化
func (o *BObject) Bencode(w io.Writer) int {
	bw, ok := w.(*bufio.Writer)
	if !ok {
		bw = bufio.NewWriter(w)
	}
	wLen := 0
	switch o.type_ {
	case BSTR:
		str, _ := o.Str()
		wLen += EncodeString(bw, str)
	case BINT:
		val, _ := o.Int()
		wLen += EncodeInt(bw, val)
	case BLIST:
		bw.WriteByte('l')
		list, _ := o.List()
		for _, elem := range list {
			wLen += elem.Bencode(bw)
		}
		bw.WriteByte('e')
		wLen += 2
	case BDICT:
		bw.WriteByte('d')
		dict, _ := o.Dict()
		for k, v := range dict {
			wLen += EncodeString(bw, k)
			wLen += v.Bencode(bw)
		}
		bw.WriteByte('e')
		wLen += 2
	}
	bw.Flush()
	return wLen
}

// string的编解码
func EncodeString(w io.Writer, val string) int {
	strLen := len(val)
	bw := bufio.NewWriter(w)
	wLen := writeDecimal(bw, strLen)
	bw.WriteByte(':')
	wLen++
	bw.WriteString(val)
	wLen += strLen
	err := bw.Flush()
	if err != nil {
		return 0
	}
	return wLen
}
func Decodestring(r io.Reader) (val string, err error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	num, len := readDecimal(br)
	if len == 0 {
		return val, ErrNum
	}
	b, err := br.ReadByte()
	if b != ':' {
		return val, ErrCol
	}
	buf := make([]byte, num)
	_, err = io.ReadAtLeast(br, buf, num)
	val = string(buf)
	return
}

func writeDecimal(w *bufio.Writer, val int) (len int) {
	if val == 0 {
		w.WriteByte('0')
		len++
		return
	}
	if val < 0 {
		w.WriteByte('-')
		len++
		val *= -1
	}
	dividend := 1
	for {
		if dividend > val {
			dividend /= 10
			break
		}
		dividend *= 10
	}
	for {
		num := byte(val / dividend)
		w.WriteByte('0' + num)
		len++
		if dividend == 1 {
			return
		}
		val %= dividend
		dividend /= 10
	}
}

// Int的编解码
func EncodeInt(w io.Writer, val int) int {
	bw := bufio.NewWriter(w)
	wLen := 0
	bw.WriteByte('i')
	wLen++
	nLen := writeDecimal(bw, val)
	wLen += nLen
	bw.WriteByte('e')
	wLen++

	err := bw.Flush()
	if err != nil {
		return 0
	}
	return wLen
}
func DecodeInt(r io.Reader) (val int, err error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	b, err := br.ReadByte()
	if b != 'i' {
		return val, ErrEpI
	}
	val, _ := readDecimal(br)
	b, err = br.ReadByte()
	if b != 'e' {
		return val, ErrEpE
	}
	return
}
