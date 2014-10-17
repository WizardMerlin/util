// Copyright 2013 Fredrik Ehnbom
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

// The binary package implements code aiding in dealing with binary data.
// The goal is to have users write as little custom binary parsing code as
// possible by focusing instead on defining the data structures in terms of
// Go structs and providing struct tags to guide the loading and saving of
// the binary data.
package binary

import (
	sb "encoding/binary"
	"fmt"
	"github.com/quarnster/util/encoding/binary/expression"
	"io"
	"math"
	"reflect"
	"unsafe"
)

type (
	// If a data type being loaded implements the validateable interface,
	// the Validate function will be called once the BinaryReader has
	// finished reading the interface, and the error if any returned from
	// this function will be what is returned from the BinaryReader's
	// ReadInterface method.
	Validateable interface {
		Validate() error
	}

	// The Reader interface gives the user a chance to perform custom
	// actions required to load specific data types.
	Reader interface {
		Read(*BinaryReader) error
	}

	// The BinaryReader uses information provided in struct tags to deal with
	// operations common when reading data from a binary file into a Go struct,
	// such as data alignment, array lengths, "if" checks and skipping of
	// uninteresting data.
	//
	// In many instances this means that no custom loading code needs to be written
	// and the user can focus on describing the data structures instead.
	//
	// For more complex needs, the Reader interface can be implemented which then
	// allows the user to write custom loading code only where it is needed.
	BinaryReader struct {
		Reader    io.ReadSeeker
		Endianess sb.ByteOrder
	}
)

var (
	LittleEndian = sb.LittleEndian
	BigEndian    = sb.BigEndian
)

func (r *BinaryReader) ReadInterface(v interface{}) error {
	if ri, ok := v.(Reader); ok {
		return ri.Read(r)
	}
	t := reflect.ValueOf(v)
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("Expected a pointer not %s", t.Kind())
	}
	v2 := t.Elem()
	switch v2.Kind() {
	case reflect.Bool:
		if d, err := r.Uint8(); err != nil {
			return err
		} else {
			v2.SetBool(bool(d != 0))
		}
	case reflect.Uint:
		if d, err := r.Uint64(); err != nil {
			return err
		} else {
			v2.SetUint(uint64(d))
		}
	case reflect.Uint64:
		if d, err := r.Uint64(); err != nil {
			return err
		} else {
			v2.SetUint(uint64(d))
		}
	case reflect.Uint32:
		if d, err := r.Uint32(); err != nil {
			return err
		} else {
			v2.SetUint(uint64(d))
		}
	case reflect.Uint16:
		if d, err := r.Uint16(); err != nil {
			return err
		} else {
			v2.SetUint(uint64(d))
		}
	case reflect.Uint8:
		if d, err := r.Uint8(); err != nil {
			return err
		} else {
			v2.SetUint(uint64(d))
		}
	case reflect.Int:
		if d, err := r.Int64(); err != nil {
			return err
		} else {
			v2.SetInt(int64(d))
		}
	case reflect.Int64:
		if d, err := r.Int64(); err != nil {
			return err
		} else {
			v2.SetInt(int64(d))
		}
	case reflect.Int32:
		if d, err := r.Int32(); err != nil {
			return err
		} else {
			v2.SetInt(int64(d))
		}
	case reflect.Int16:
		if d, err := r.Int16(); err != nil {
			return err
		} else {
			v2.SetInt(int64(d))
		}
	case reflect.Int8:
		if d, err := r.Int8(); err != nil {
			return err
		} else {
			v2.SetInt(int64(d))
		}
	case reflect.Float32:
		if f, err := r.Float32(); err != nil {
			return err
		} else {
			v2.SetFloat(float64(f))
		}
	case reflect.Float64:
		if f, err := r.Float64(); err != nil {
			return err
		} else {
			v2.SetFloat(f)
		}
	case reflect.Array:
		for i := 0; i < v2.Len(); i++ {
			if err := r.ReadInterface(v2.Index(i).Addr().Interface()); err != nil {
				return err
			}
		}
	case reflect.Slice:
		for i := 0; i < v2.Len(); i++ {
			if err := r.ReadInterface(v2.Index(i).Addr().Interface()); err != nil {
				return err
			}
		}
	case reflect.String:
		var data []byte
		var max = math.MaxInt32

		for i := 0; i < max; i++ {
			if u, err := r.Uint8(); err != nil {
				return err
			} else if u == '\u0000' {
				break
			} else {
				data = append(data, u)
			}
		}
		v2.SetString(string(data))
	case reflect.Struct:
		for i := 0; i < v2.NumField(); i++ {
			var (
				f    = v2.Field(i)
				f2   = v2.Type().Field(i)
				size = -1
				err  error
			)
			if fi := f2.Tag.Get("if"); fi != "" {
				var e expression.EXPRESSION
				if !e.Parse(fi) {
					return e.Error()
				} else if ev, err := expression.Eval(&v2, e.RootNode()); err != nil {
					return err
				} else if ev == 0 {
					continue
				}
			}
			if l := f2.Tag.Get("skip"); l != "" {
				var e expression.EXPRESSION
				if !e.Parse(l) {
					return e.Error()
				} else if ev, err := expression.Eval(&v2, e.RootNode()); err != nil {
					return err
				} else if _, err := r.Seek(int64(ev), 1); err != nil {
					return err
				}
			}

			if l := f2.Tag.Get("length"); l != "" {
				switch l {
				case "uint8":
					if s, err := r.Uint8(); err != nil {
						return err
					} else {
						size = int(s)
					}
				case "uint16":
					if s, err := r.Uint16(); err != nil {
						return err
					} else {
						size = int(s)
					}
				case "uint32":
					if s, err := r.Uint32(); err != nil {
						return err
					} else {
						size = int(s)
					}
				case "uint64":
					if s, err := r.Uint64(); err != nil {
						return err
					} else {
						size = int(s)
					}
				default:
					var e expression.EXPRESSION
					if !e.Parse(l) {
						return e.Error()
					} else if ev, err := expression.Eval(&v2, e.RootNode()); err != nil {
						return err
					} else {
						size = ev
					}
				}
			}

			switch f.Type().Kind() {
			case reflect.String:
				var data []byte
				if size >= 0 {
					if data, err = r.Read(size); err != nil {
						return err
					}
					for i, v := range data {
						if v == '\u0000' {
							data = data[:i]
							break
						}
					}
				} else {
					var max = math.MaxInt32
					if m := f2.Tag.Get("max"); m != "" {
						var e expression.EXPRESSION
						if !e.Parse(m) {
							return e.Error()
						} else if ev, err := expression.Eval(&v2, e.RootNode()); err != nil {
							return err
						} else {
							max = ev
						}
					}

					for i := 0; i < max; i++ {
						if u, err := r.Uint8(); err != nil {
							return err
						} else if u == '\u0000' {
							size = i + 1
							break
						} else {
							data = append(data, u)
						}
					}
				}
				f.SetString(string(data))
			case reflect.Slice:
				if size == -1 {
					return fmt.Errorf("SliceHeader require a known length, %+v", v)
				}
				if f.Type().Elem().Kind() == reflect.Int8 {
					if b, err := r.Read(size); err != nil {
						return err
					} else {
						f.Set(reflect.ValueOf(b))
					}
				} else {
					var v3 = reflect.MakeSlice(f.Type(), size, size)
					for i := 0; i < size; i++ {
						if err = r.ReadInterface(v3.Index(i).Addr().Interface()); err != nil {
							return err
						}
					}
					f.Set(v3)
				}
			default:
				if err := r.ReadInterface(f.Addr().Interface()); err != nil {
					return err
				} else {
					size = int(f.Type().Size())
				}
			}

			if al := f2.Tag.Get("align"); al != "" {
				var (
					e     expression.EXPRESSION
					align int
					seek  int
				)
				if !e.Parse(al) {
					return e.Error()
				} else if ev, err := expression.Eval(&v2, e.RootNode()); err != nil {
					return err
				} else {
					align = ev
				}
				if align < size {
					seek = ((size + (align - 1)) &^ (align - 1)) - size
				} else if align > size {
					seek = align - size
				}
				if seek > 0 {
					if _, err := r.Seek(int64(seek), 1); err != nil {
						return err
					}
				}
			}
		}
	default:
		return fmt.Errorf("Don't know how to read type %s", v2.Kind())
	}
	if val, ok := v.(Validateable); ok {
		return val.Validate()
	}
	return nil
}

func (r *BinaryReader) Seek(offset int64, whence int) (int64, error) {
	return r.Reader.Seek(offset, whence)
}

func (r *BinaryReader) Read(size int) ([]byte, error) {
	data := make([]byte, size)
	if size == 0 {
		return data, nil
	}
	if n, err := r.Reader.Read(data); err != nil {
		return nil, err
	} else if n != len(data) {
		return nil, fmt.Errorf("Didn't read the expected number of bytes")
	}
	return data, nil
}

func (r *BinaryReader) Uint64() (uint64, error) {
	if data, err := r.Read(8); err != nil {
		return 0, err
	} else {
		return r.Endianess.Uint64(data), nil
	}
}

func (r *BinaryReader) Uint32() (uint32, error) {
	if data, err := r.Read(4); err != nil {
		return 0, err
	} else {
		return r.Endianess.Uint32(data), nil
	}
}

func (r *BinaryReader) Uint16() (uint16, error) {
	if data, err := r.Read(2); err != nil {
		return 0, err
	} else {
		return r.Endianess.Uint16(data), nil
	}
}

func (r *BinaryReader) Uint8() (uint8, error) {
	if data, err := r.Read(1); err != nil {
		return 0, err
	} else {
		return uint8(data[0]), nil
	}
}

func (r *BinaryReader) Int64() (int64, error) {
	if data, err := r.Uint64(); err != nil {
		return 0, err
	} else {
		return int64(data), nil
	}
}

func (r *BinaryReader) Int32() (int32, error) {
	if data, err := r.Uint32(); err != nil {
		return 0, err
	} else {
		return int32(data), nil
	}
}

func (r *BinaryReader) Int16() (int16, error) {
	if data, err := r.Uint16(); err != nil {
		return 0, err
	} else {
		return int16(data), nil
	}
}

func (r *BinaryReader) Int8() (int8, error) {
	if data, err := r.Read(1); err != nil {
		return 0, err
	} else {
		return int8(data[0]), nil
	}
}

func (r *BinaryReader) Float32() (float32, error) {
	if i32, err := r.Int32(); err != nil {
		return 0, err
	} else {
		f32 := *(*float32)(unsafe.Pointer(&i32))
		return f32, nil
	}
}

func (r *BinaryReader) Float64() (float64, error) {
	if i64, err := r.Int64(); err != nil {
		return 0, err
	} else {
		f64 := *(*float64)(unsafe.Pointer(&i64))
		return f64, nil
	}
}
