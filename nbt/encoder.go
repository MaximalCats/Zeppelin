package nbt

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"unsafe"
)

type Encoder struct {
	w                         io.Writer
	dontWriteRootCompoundName bool
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (e *Encoder) Encode(name string, v any) error {
	if err := e.writeByte(Compound); err != nil {
		return err
	}
	if !e.dontWriteRootCompoundName {
		if err := e.writeString(name); err != nil {
			return err
		}
	}
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Struct:
		fmt.Println("encoding root struct compound")
		return e.encodeCompoundStruct(val)
	case reflect.Map:
		fmt.Println("encoding root map compound")
		return e.encodeCompoundMap(val)
	default:
		return fmt.Errorf("Encode expects map/struct, not %s", val.Type())
	}
}

func (e *Encoder) WriteRootName(val bool) {
	e.dontWriteRootCompoundName = !val
}

func (e *Encoder) writeBytes(b ...byte) error {
	_, err := e.w.Write(b)
	return err
}

func (e *Encoder) writeByte(b int8) error {
	return e.writeBytes(byte(b))
}

func (e *Encoder) writeShort(s int16) error {
	return e.writeBytes(
		byte(s>>8),
		byte(s),
	)
}

func (e *Encoder) writeInt(i int32) error {
	return e.writeBytes(
		byte(i>>24),
		byte(i>>16),
		byte(i>>8),
		byte(i),
	)
}

func (e *Encoder) writeLong(l int64) error {
	return e.writeBytes(
		byte(l>>56),
		byte(l>>48),
		byte(l>>40),
		byte(l>>32),
		byte(l>>24),
		byte(l>>16),
		byte(l>>8),
		byte(l),
	)
}

func (e *Encoder) writeFloat(f float32) error {
	return e.writeInt(int32(math.Float32bits(f)))
}

func (e *Encoder) writeDouble(d float64) error {
	return e.writeLong(int64(math.Float64bits(d)))
}

func (e *Encoder) writeByteArray(ba []int8) error {
	if err := e.writeInt(int32(len(ba))); err != nil {
		return err
	}
	return e.writeBytes(*(*[]byte)(unsafe.Pointer(&ba))...)
}

func (e *Encoder) writeIntArray(il []int32) error {
	if err := e.writeInt(int32(len(il))); err != nil {
		return err
	}
	return binary.Write(e.w, binary.BigEndian, il)
}

func (e *Encoder) writeLongArray(ll []int64) error {
	if err := e.writeInt(int32(len(ll))); err != nil {
		return err
	}
	return binary.Write(e.w, binary.BigEndian, ll)
}

func (e *Encoder) writeString(s string) error {
	if err := e.writeShort(int16(len(s))); err != nil {
		return err
	}
	return e.writeBytes([]byte(s)...)
}

func (e *Encoder) encodeCompoundStruct(val reflect.Value) error {
	for i := 0; i < val.NumField(); i++ {
		tf := val.Type().Field(i)
		f := val.Field(i)
		if !tf.IsExported() {
			fmt.Println(tf.Name, "unexported")
			continue
		}
		name := tf.Name
		if v, ok := tf.Tag.Lookup("nbt"); ok {
			name = v
		}

		switch f.Kind() {
		case reflect.Bool:
			fmt.Println("encoding bool", name, f.Bool())
			if err := e.writeByte(Byte); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			b := f.Bool()
			if err := e.writeByte(*(*int8)(unsafe.Pointer(&b))); err != nil {
				return err
			}
		case reflect.Int8, reflect.Uint8:
			fmt.Println("encoding byte", name)
			if err := e.writeByte(Byte); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			if f.CanUint() {
				if err := e.writeByte(int8(f.Uint())); err != nil {
					return err
				}
			} else {
				if err := e.writeByte(int8(f.Int())); err != nil {
					return err
				}
			}
		case reflect.Int16, reflect.Uint16:
			fmt.Println("encoding short", name)
			if err := e.writeByte(Short); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			if f.CanUint() {
				if err := e.writeShort(int16(f.Uint())); err != nil {
					return err
				}
			} else {
				if err := e.writeShort(int16(f.Int())); err != nil {
					return err
				}
			}
		case reflect.Int32, reflect.Uint32:
			fmt.Println("encoding int", name)
			if err := e.writeByte(Int); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			if f.CanUint() {
				if err := e.writeInt(int32(f.Uint())); err != nil {
					return err
				}
			} else {
				if err := e.writeInt(int32(f.Int())); err != nil {
					return err
				}
			}
		case reflect.Int64, reflect.Uint64:
			fmt.Println("encoding long", name)
			if err := e.writeByte(Long); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			if f.CanUint() {
				if err := e.writeLong(int64(f.Uint())); err != nil {
					return err
				}
			} else {
				if err := e.writeLong(int64(f.Int())); err != nil {
					return err
				}
			}
		case reflect.Float32:
			fmt.Println("encoding float", name)
			if err := e.writeByte(Float); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			if err := e.writeFloat(float32(f.Float())); err != nil {
				return err
			}
		case reflect.Float64:
			fmt.Println("encoding double", name)
			if err := e.writeByte(Double); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			if err := e.writeDouble(f.Float()); err != nil {
				return err
			}
		case reflect.String:
			fmt.Println("encoding string", name, f.String())
			if err := e.writeByte(String); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			if err := e.writeString(f.String()); err != nil {
				return err
			}
		case reflect.Slice, reflect.Array:
			switch f.Type().Elem().Kind() {
			case reflect.Uint8, reflect.Int8:
				fmt.Println("encoding byte array", name)
				if err := e.writeByte(ByteArray); err != nil {
					return err
				}
				if err := e.writeString(name); err != nil {
					return err
				}
				if err := e.writeBytes(*(*[]byte)(f.UnsafePointer())...); err != nil {
					return err
				}
			case reflect.Uint32, reflect.Int32:
				fmt.Println("encoding int array", name)
				if err := e.writeByte(IntArray); err != nil {
					return err
				}
				if err := e.writeString(name); err != nil {
					return err
				}
				if err := e.writeIntArray(*(*[]int32)(f.UnsafePointer())); err != nil {
					return err
				}
			case reflect.Uint64, reflect.Int64:
				fmt.Println("encoding long array")
				if err := e.writeByte(LongArray); err != nil {
					return err
				}
				if err := e.writeString(name); err != nil {
					return err
				}
				if err := e.writeLongArray(*(*[]int64)(f.UnsafePointer())); err != nil {
					return err
				}
			default:
				fmt.Println("encoding list", name)
				if err := e.writeByte(List); err != nil {
					return err
				}
				if err := e.writeString(name); err != nil {
					return err
				}
				if err := e.encodeList(f); err != nil {
					return err
				}
			}
		case reflect.Struct:
			fmt.Println("encoding struct compound", name)
			if err := e.writeByte(Compound); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			if err := e.encodeCompoundStruct(f); err != nil {
				return err
			}
		case reflect.Map:
			fmt.Println("encoding map compound", name)
			if err := e.writeByte(Compound); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			if err := e.encodeCompoundMap(f); err != nil {
				return err
			}
		}

	}
	fmt.Println("struct compound done")
	return e.writeByte(0)
}

func (e *Encoder) encodeCompoundMap(val reflect.Value) error {
	for _, key := range val.MapKeys() {
		f := val.MapIndex(key)
		name := key.String()

		switch f.Kind() {
		case reflect.Bool:
			if err := e.writeByte(Byte); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			b := f.Bool()
			if err := e.writeByte(*(*int8)(unsafe.Pointer(&b))); err != nil {
				return err
			}
		case reflect.Int8, reflect.Uint8:
			if err := e.writeByte(Byte); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			if f.CanUint() {
				if err := e.writeByte(int8(f.Uint())); err != nil {
					return err
				}
			} else {
				if err := e.writeByte(int8(f.Int())); err != nil {
					return err
				}
			}
		case reflect.Int16, reflect.Uint16:
			if err := e.writeByte(Short); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			if f.CanUint() {
				if err := e.writeShort(int16(f.Uint())); err != nil {
					return err
				}
			} else {
				if err := e.writeShort(int16(f.Int())); err != nil {
					return err
				}
			}
		case reflect.Int32, reflect.Uint32:
			if err := e.writeByte(Int); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			if f.CanUint() {
				if err := e.writeInt(int32(f.Uint())); err != nil {
					return err
				}
			} else {
				if err := e.writeInt(int32(f.Int())); err != nil {
					return err
				}
			}
		case reflect.Int64, reflect.Uint64:
			if err := e.writeByte(Long); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			if f.CanUint() {
				if err := e.writeLong(int64(f.Uint())); err != nil {
					return err
				}
			} else {
				if err := e.writeLong(int64(f.Int())); err != nil {
					return err
				}
			}
		case reflect.Float32:
			if err := e.writeByte(Float); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			if err := e.writeFloat(float32(f.Float())); err != nil {
				return err
			}
		case reflect.Float64:
			if err := e.writeByte(Double); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			if err := e.writeDouble(f.Float()); err != nil {
				return err
			}
		case reflect.String:
			if err := e.writeByte(String); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			if err := e.writeString(f.String()); err != nil {
				return err
			}
		case reflect.Slice, reflect.Array:
			switch f.Type().Elem().Kind() {
			case reflect.Uint8, reflect.Int8:
				if err := e.writeByte(ByteArray); err != nil {
					return err
				}
				if err := e.writeString(name); err != nil {
					return err
				}
				if err := e.writeBytes(*(*[]byte)(f.UnsafePointer())...); err != nil {
					return err
				}
			case reflect.Uint32, reflect.Int32:
				if err := e.writeByte(IntArray); err != nil {
					return err
				}
				if err := e.writeString(name); err != nil {
					return err
				}
				if err := e.writeIntArray(*(*[]int32)(f.UnsafePointer())); err != nil {
					return err
				}
			case reflect.Uint64, reflect.Int64:
				if err := e.writeByte(LongArray); err != nil {
					return err
				}
				if err := e.writeString(name); err != nil {
					return err
				}
				if err := e.writeLongArray(*(*[]int64)(f.UnsafePointer())); err != nil {
					return err
				}
			default:
				if err := e.writeByte(List); err != nil {
					return err
				}
				if err := e.writeString(name); err != nil {
					return err
				}
				if err := e.encodeList(f); err != nil {
					return err
				}
			}
		case reflect.Struct:
			if err := e.writeByte(Compound); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			if err := e.encodeCompoundStruct(f); err != nil {
				return err
			}
		case reflect.Map:
			if err := e.writeByte(Compound); err != nil {
				return err
			}
			if err := e.writeString(name); err != nil {
				return err
			}
			if err := e.encodeCompoundMap(f); err != nil {
				return err
			}
		}

	}
	return e.writeByte(0)
}

func (e *Encoder) encodeList(val reflect.Value) error {
	fmt.Println("list type", e.tagFor(val.Type().Elem()))
	if err := e.writeByte(e.tagFor(val.Type().Elem())); err != nil {
		return err
	}

	fmt.Println("list len", val.Len())

	if err := e.writeInt(int32(val.Len())); err != nil {
		return err
	}

	for i := 0; i < val.Len(); i++ {
		f := val.Index(i)
		switch val.Type().Elem().Kind() {
		case reflect.Bool:
			fmt.Println("writing bool", i)
			b := f.Bool()
			if err := e.writeByte(*(*int8)(unsafe.Pointer(&b))); err != nil {
				return err
			}
		case reflect.Int8:
			fmt.Println("writing byte", i)
			if err := e.writeByte(int8(f.Int())); err != nil {
				return err
			}
		case reflect.Uint8:
			fmt.Println("writing byte", i)
			if err := e.writeByte(int8(f.Uint())); err != nil {
				return err
			}
		case reflect.Int16:
			fmt.Println("writing short", i)
			if err := e.writeShort(int16(f.Int())); err != nil {
				return err
			}
		case reflect.Uint16:
			fmt.Println("writing short", i)
			if err := e.writeShort(int16(f.Uint())); err != nil {
				return err
			}
		case reflect.Int32:
			fmt.Println("writing int", i)
			if err := e.writeInt(int32(f.Int())); err != nil {
				return err
			}
		case reflect.Uint32:
			fmt.Println("writing int", i)
			if err := e.writeInt(int32(f.Uint())); err != nil {
				return err
			}
		case reflect.Int64:
			fmt.Println("writing long", i)
			if err := e.writeLong(f.Int()); err != nil {
				return err
			}
		case reflect.Uint64:
			fmt.Println("writing long", i)
			if err := e.writeLong(int64(f.Uint())); err != nil {
				return err
			}
		case reflect.String:
			fmt.Println("writing string", i, f.String())
			if err := e.writeString(f.String()); err != nil {
				return err
			}
		case reflect.Slice, reflect.Array:
			switch f.Type().Elem().Kind() {
			case reflect.Uint8, reflect.Int8:
				fmt.Println("writing byte array", i)
				if err := e.writeBytes(*(*[]byte)(f.UnsafePointer())...); err != nil {
					return err
				}
			case reflect.Uint32, reflect.Int32:
				fmt.Println("writing int array", i)
				if err := e.writeIntArray(*(*[]int32)(f.UnsafePointer())); err != nil {
					return err
				}
			case reflect.Uint64, reflect.Int64:
				fmt.Println("writing long array", i)
				if err := e.writeLongArray(*(*[]int64)(f.UnsafePointer())); err != nil {
					return err
				}
			default:
				fmt.Println("writing list", i)
				if err := e.encodeList(f); err != nil {
					return err
				}
			}
		case reflect.Float32:
			fmt.Println("writing float", i)
			if err := e.writeFloat(float32(f.Float())); err != nil {
				return err
			}
		case reflect.Float64:
			fmt.Println("writing double", i)
			if err := e.writeDouble(f.Float()); err != nil {
				return err
			}
		case reflect.Struct:
			fmt.Println("writing struct compound", i)
			if err := e.encodeCompoundStruct(f); err != nil {
				return err
			}
		case reflect.Map:
			fmt.Println("writing map compound", i)
			if err := e.encodeCompoundMap(f); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *Encoder) tagFor(typ reflect.Type) int8 {
	switch typ.Kind() {
	case reflect.Uint8, reflect.Int8, reflect.Bool:
		return Byte
	case reflect.Uint16, reflect.Int16:
		return Short
	case reflect.Uint32, reflect.Int32:
		return Int
	case reflect.Uint64, reflect.Int64:
		return Long
	case reflect.Float32:
		return Float
	case reflect.Float64:
		return Double
	case reflect.Struct, reflect.Map:
		return Compound
	case reflect.String:
		return String
	case reflect.Slice, reflect.Array:
		{
			switch typ.Elem().Kind() {
			case reflect.Uint8, reflect.Int8:
				return ByteArray
			case reflect.Uint32, reflect.Int32:
				return IntArray
			case reflect.Uint64, reflect.Int64:
				return LongArray
			default:
				return List
			}
		}
	default:
		return 0
	}
}
