package nbt

// Visitor defines the interface for visiting NBT structures in a streaming fashion.
// This allows processing NBT data without loading it entirely into memory.
type Visitor interface {
	// VisitByte is called for TAG_Byte.
	VisitByte(value int8) error

	// VisitShort is called for TAG_Short.
	VisitShort(value int16) error

	// VisitInt is called for TAG_Int.
	VisitInt(value int32) error

	// VisitLong is called for TAG_Long.
	VisitLong(value int64) error

	// VisitFloat is called for TAG_Float.
	VisitFloat(value float32) error

	// VisitDouble is called for TAG_Double.
	VisitDouble(value float64) error

	// VisitByteArray is called for TAG_Byte_Array.
	VisitByteArray(value []byte) error

	// VisitString is called for TAG_String.
	VisitString(value string) error

	// VisitIntArray is called for TAG_Int_Array.
	VisitIntArray(value []int32) error

	// VisitLongArray is called for TAG_Long_Array.
	VisitLongArray(value []int64) error

	// VisitListStart is called at the start of a TAG_List.
	// Returns a new Visitor for list elements, or nil to skip.
	VisitListStart(elementType byte, length int) (Visitor, error)

	// VisitListEnd is called at the end of a TAG_List.
	VisitListEnd() error

	// VisitCompoundStart is called at the start of a TAG_Compound.
	// Returns a new Visitor for compound entries, or nil to skip.
	VisitCompoundStart() (Visitor, error)

	// VisitCompoundEntry is called for each entry in a TAG_Compound.
	// Returns a new Visitor for the entry value, or nil to skip.
	VisitCompoundEntry(name string, tagType byte) (Visitor, error)

	// VisitCompoundEnd is called at the end of a TAG_Compound.
	VisitCompoundEnd() error

	// VisitEnd is called for TAG_End.
	VisitEnd() error
}

// BaseVisitor provides default implementations for the Visitor interface.
// Embed this in custom visitors to only override specific methods.
type BaseVisitor struct{}

func (BaseVisitor) VisitByte(int8) error         { return nil }
func (BaseVisitor) VisitShort(int16) error       { return nil }
func (BaseVisitor) VisitInt(int32) error         { return nil }
func (BaseVisitor) VisitLong(int64) error        { return nil }
func (BaseVisitor) VisitFloat(float32) error     { return nil }
func (BaseVisitor) VisitDouble(float64) error    { return nil }
func (BaseVisitor) VisitByteArray([]byte) error  { return nil }
func (BaseVisitor) VisitString(string) error     { return nil }
func (BaseVisitor) VisitIntArray([]int32) error  { return nil }
func (BaseVisitor) VisitLongArray([]int64) error { return nil }
func (BaseVisitor) VisitListStart(byte, int) (Visitor, error) {
	return nil, nil // Skip by default
}
func (BaseVisitor) VisitListEnd() error { return nil }
func (BaseVisitor) VisitCompoundStart() (Visitor, error) {
	return nil, nil // Skip by default
}
func (BaseVisitor) VisitCompoundEntry(string, byte) (Visitor, error) {
	return nil, nil // Skip by default
}
func (BaseVisitor) VisitCompoundEnd() error { return nil }
func (BaseVisitor) VisitEnd() error         { return nil }

// AcceptVisitor visits an NBT tag with the given visitor.
func AcceptVisitor(tag Tag, v Visitor) error {
	if v == nil {
		return nil
	}

	switch t := tag.(type) {
	case Byte:
		return v.VisitByte(int8(t))
	case Short:
		return v.VisitShort(int16(t))
	case Int:
		return v.VisitInt(int32(t))
	case Long:
		return v.VisitLong(int64(t))
	case Float:
		return v.VisitFloat(float32(t))
	case Double:
		return v.VisitDouble(float64(t))
	case ByteArray:
		return v.VisitByteArray([]byte(t))
	case String:
		return v.VisitString(string(t))
	case IntArray:
		return v.VisitIntArray([]int32(t))
	case LongArray:
		return v.VisitLongArray([]int64(t))
	case List:
		return acceptListVisitor(t, v)
	case Compound:
		return acceptCompoundVisitor(t, v)
	case End:
		return v.VisitEnd()
	default:
		return nil
	}
}

func acceptListVisitor(list List, v Visitor) error {
	elemVisitor, err := v.VisitListStart(list.ElementType, len(list.Elements))
	if err != nil {
		return err
	}

	if elemVisitor != nil {
		for _, elem := range list.Elements {
			if err := AcceptVisitor(elem, elemVisitor); err != nil {
				return err
			}
		}
	}

	return v.VisitListEnd()
}

func acceptCompoundVisitor(compound Compound, v Visitor) error {
	compoundVisitor, err := v.VisitCompoundStart()
	if err != nil {
		return err
	}

	if compoundVisitor != nil {
		for name, tag := range compound {
			entryVisitor, err := compoundVisitor.VisitCompoundEntry(name, tag.ID())
			if err != nil {
				return err
			}
			if entryVisitor != nil {
				if err := AcceptVisitor(tag, entryVisitor); err != nil {
					return err
				}
			}
		}
	}

	return v.VisitCompoundEnd()
}

// VisitReader reads NBT data using a visitor without fully loading into memory.
// This is useful for processing large NBT files.
func VisitReader(r *Reader, v Visitor, network bool) error {
	// Read tag type
	tagType, err := r.readByte()
	if err != nil {
		return err
	}

	if tagType == TagEnd {
		return v.VisitEnd()
	}

	// Skip root name for file format
	if !network {
		if _, err := r.readString(); err != nil {
			return err
		}
	}

	return visitTagPayload(r, tagType, v)
}

func visitTagPayload(r *Reader, tagType byte, v Visitor) error {
	if v == nil {
		// Skip this tag
		return skipTagPayload(r, tagType)
	}

	switch tagType {
	case TagEnd:
		return v.VisitEnd()

	case TagByte:
		val, err := r.readByte()
		if err != nil {
			return err
		}
		return v.VisitByte(int8(val))

	case TagShort:
		val, err := r.readShort()
		if err != nil {
			return err
		}
		return v.VisitShort(val)

	case TagInt:
		val, err := r.readInt()
		if err != nil {
			return err
		}
		return v.VisitInt(val)

	case TagLong:
		val, err := r.readLong()
		if err != nil {
			return err
		}
		return v.VisitLong(val)

	case TagFloat:
		val, err := r.readFloat()
		if err != nil {
			return err
		}
		return v.VisitFloat(val)

	case TagDouble:
		val, err := r.readDouble()
		if err != nil {
			return err
		}
		return v.VisitDouble(val)

	case TagByteArray:
		length, err := r.readInt()
		if err != nil {
			return err
		}
		data := make([]byte, length)
		if err := r.readFull(data); err != nil {
			return err
		}
		return v.VisitByteArray(data)

	case TagString:
		val, err := r.readString()
		if err != nil {
			return err
		}
		return v.VisitString(val)

	case TagList:
		return visitList(r, v)

	case TagCompound:
		return visitCompound(r, v)

	case TagIntArray:
		length, err := r.readInt()
		if err != nil {
			return err
		}
		data := make([]int32, length)
		for i := range data {
			data[i], err = r.readInt()
			if err != nil {
				return err
			}
		}
		return v.VisitIntArray(data)

	case TagLongArray:
		length, err := r.readInt()
		if err != nil {
			return err
		}
		data := make([]int64, length)
		for i := range data {
			data[i], err = r.readLong()
			if err != nil {
				return err
			}
		}
		return v.VisitLongArray(data)

	default:
		return nil
	}
}

func visitList(r *Reader, v Visitor) error {
	elemType, err := r.readByte()
	if err != nil {
		return err
	}

	length, err := r.readInt()
	if err != nil {
		return err
	}

	elemVisitor, err := v.VisitListStart(elemType, int(length))
	if err != nil {
		return err
	}

	for range length {
		if err := visitTagPayload(r, elemType, elemVisitor); err != nil {
			return err
		}
	}

	return v.VisitListEnd()
}

func visitCompound(r *Reader, v Visitor) error {
	compoundVisitor, err := v.VisitCompoundStart()
	if err != nil {
		return err
	}

	for {
		tagType, err := r.readByte()
		if err != nil {
			return err
		}

		if tagType == TagEnd {
			break
		}

		name, err := r.readString()
		if err != nil {
			return err
		}

		var entryVisitor Visitor
		if compoundVisitor != nil {
			entryVisitor, err = compoundVisitor.VisitCompoundEntry(name, tagType)
			if err != nil {
				return err
			}
		}

		if err := visitTagPayload(r, tagType, entryVisitor); err != nil {
			return err
		}
	}

	return v.VisitCompoundEnd()
}

func skipTagPayload(r *Reader, tagType byte) error {
	switch tagType {
	case TagEnd:
		return nil
	case TagByte:
		_, err := r.readByte()
		return err
	case TagShort:
		_, err := r.readShort()
		return err
	case TagInt:
		_, err := r.readInt()
		return err
	case TagLong:
		_, err := r.readLong()
		return err
	case TagFloat:
		_, err := r.readFloat()
		return err
	case TagDouble:
		_, err := r.readDouble()
		return err
	case TagByteArray:
		length, err := r.readInt()
		if err != nil {
			return err
		}
		data := make([]byte, length)
		return r.readFull(data)
	case TagString:
		_, err := r.readString()
		return err
	case TagList:
		elemType, err := r.readByte()
		if err != nil {
			return err
		}
		length, err := r.readInt()
		if err != nil {
			return err
		}
		for range length {
			if err := skipTagPayload(r, elemType); err != nil {
				return err
			}
		}
		return nil
	case TagCompound:
		for {
			entryType, err := r.readByte()
			if err != nil {
				return err
			}
			if entryType == TagEnd {
				return nil
			}
			if _, err := r.readString(); err != nil {
				return err
			}
			if err := skipTagPayload(r, entryType); err != nil {
				return err
			}
		}
	case TagIntArray:
		length, err := r.readInt()
		if err != nil {
			return err
		}
		for range length {
			if _, err := r.readInt(); err != nil {
				return err
			}
		}
		return nil
	case TagLongArray:
		length, err := r.readInt()
		if err != nil {
			return err
		}
		for range length {
			if _, err := r.readLong(); err != nil {
				return err
			}
		}
		return nil
	default:
		return nil
	}
}
