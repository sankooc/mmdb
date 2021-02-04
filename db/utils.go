package db

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"hash/crc32"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

var requestIDLock sync.Mutex
var requestSeed = uint32(4670)

func errSReply() bson.M {
	return bson.M{"ok": 0}
}

func okReply() bson.M {
	return bson.M{"ok": 1}
}

func markOk(msg bson.D) bson.D {
	return append(msg, bson.E{"ok", 1})
}

func newRequest() uint32 {
	requestIDLock.Lock()
	defer requestIDLock.Unlock()
	requestSeed++
	return requestSeed
}

func Debug(fmt string, args ...interface{}) {
	ld := os.Getenv("DEBUG")
	if ld != "" {
		log.Printf(fmt, args...)
	}
}

func readHeaderData(reader io.Reader) (*MongoMessage, error) {
	Length := readUInt32(reader)
	if Length < 16 {
		return nil, fmt.Errorf("data_length_error %d", Length)
	}
	RequestID := readUInt32(reader)
	ResponseTo := readUInt32(reader)
	OpCode := readUInt32(reader)
	Data := make([]byte, Length-16)
	reader.Read(Data)
	return &MongoMessage{Length, RequestID, ResponseTo, OpCode, Data}, nil
}

func readUInt32(reader io.Reader) uint32 {
	bit := make([]byte, 4)
	reader.Read(bit)
	return binary.LittleEndian.Uint32(bit)
}

func readUInt8(reader *bytes.Reader) byte {
	//buf := make([]byte, 1)
	//reader.Read(buf)
	//return buf[0]
	c,_ := reader.ReadByte()
	return c
}

func resetReader(reader *bytes.Reader, size int) {
	bk := 0
	for bk < size {
		reader.UnreadByte()
		bk += 1
	}
}

func unmarshalDoc(reader *bytes.Reader, out interface{}) (int, error) {
	size := readUInt32(reader)
	Debug("doc size %d \r\n", size)
	resetReader(reader, 4)
	buf := make([]byte, size)
	reader.Read(buf)
	if err := bson.Unmarshal(buf, out); err != nil {
		return int(size), err
	}
	return int(size), nil
}

func readString(reader *bytes.Reader) string {
	var offset = 0
	for true {
		b, _ := reader.ReadByte()
		offset += 1
		if b == 0 {
			break
		}
	}
	buf := make([]byte, offset-1)
	resetReader(reader, offset)
	reader.Read(buf)
	reader.ReadByte()
	return string(buf)
}

func writeUint32(writer io.Writer, num uint32) error {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, num)
	_, err := writer.Write(b)
	return err
}

func writeUint64(writer io.Writer, num uint64) error {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, num)
	_, err := writer.Write(b)
	return err
}

func parseSection(reader *bytes.Reader) ([]Section, []Section) {
	list := make([]Section, 0)
	meta := make([]Section, 0)
	for true {
		if reader.Len() <= 4 {
			Debug("ext %d \r\n", reader.Len())
			break
		}
		btype := readUInt8(reader)
		Debug("btype %d \r\n", btype)
		if btype == 1 {
			leftsize := reader.Len()
			size := int(readUInt32(reader))
			leftsize -= size
			Debug("size %d \r\n", size)
			doc := readString(reader)
			Debug("doc %s \r\n", doc)
			// for
			for reader.Len() > leftsize {
				section := make(bson.M)
				unmarshalDoc(reader, &section)
				list = append(list, section)
			}
		} else {
			section := make(bson.M)
			unmarshalDoc(reader, &section)
			meta = append(meta, section)
		}
	}
	return meta, list
}

func handle2013(req *M2013, engine MongoEngine) *M2013Reply {
	msg := &M2013Reply{Message: req.Message, FlagBits: 0 }
	msg.Message.ResponseTo = req.Message.RequestID
	msg.Message.RequestID = newRequest()

	if req.Meta != nil {
		section := req.Meta[0].(bson.M)
		ns := section["$db"]
		if section["whatsmyuri"] != nil {
			msg.sections = []Section{bson.M{"you": "127.0.0.1:56163", "ok": 1}}
		} else if section["buildinfo"] != nil {
			msg.sections = []Section{BUILDINFO}
		} else if section["getFreeMonitoringStatus"] != nil {
			msg.sections = []Section{bson.M{"state": "undecided", "ok": 1}} //TODO
		} else if section["endSessions"] != nil {
			msg.sections = []Section{Section(okReply())}
		} else if section["listDatabases"] != nil {
			msg.sections = []Section{Section(engine.listDatabases())}
		} else if section["listCollections"] != nil {
			dbname := fmt.Sprintf("%v", section["$db"])
			rs := engine.listCollections(section, dbname)
			msg.sections = []Section{rs}
		} else if section["isMaster"] != nil {
			msg.sections = []Section{ISMASTER}
		} else if section["replSetGetStatus"] != nil {
			msg.sections = []Section{bson.M{"ok": 1}}
		} else if section["getLog"] == "startupWarnings" {
			msg.sections = []Section{WARNING}
		} else if section["ismaster"] != nil {
			msg.sections = []Section{ISMASTER}
		} else if  ns != nil {
			dbname := section["$db"].(string)
			Debug(" $db [%v] dbname\r\n", dbname)
			if section["find"] != nil {
				col := section["find"]
				filter := section["filter"]
				rs := engine.query(dbname, col.(string), filter.(bson.M))
				Debug("query rs [%v]", rs)
				msg.sections = []Section{rs}
			} else if section["insert"] != nil {
				col := section["insert"].(string)
				var docs bson.A
				if len(req.Sections) > 0 {
					docs = make([]interface{}, len(req.Sections))
					for i, v := range req.Sections {
						docs[i] = v
					}
				} else {
					docs = section["documents"].(bson.A)
				}
				rs := engine.insert(dbname, col, docs)
				msg.sections = []Section{rs}
			} else if section["delete"] != nil {
				col := section["delete"].(string)
				if len(req.Sections) > 0 {
					rs := engine.delete(dbname, col, req.Sections[0].(bson.M))
					msg.sections = []Section{rs}
				}
			} else if section["update"] != nil {
				col := section["update"].(string)
				if len(req.Sections) > 0 {
					rs := engine.update(dbname, col, req.Sections[0].(bson.M))
					msg.sections = []Section{rs}
				}
			}
		} else {
			//msg.sections = []Section{}
			Debug("unknown [%v] cmd", section)
			msg.sections = []Section{markOk(nil)}
		}
		return msg
	}
	return msg
}

func handle2004(req *M2004, engine MongoEngine) *M2004Reply{
	cmd, _:= req.Command()

	msg := &M2004Reply{Message: req.Message}
	msg.Message.ResponseTo = req.Message.RequestID
	msg.Message.RequestID = newRequest()
	msg.Message.OpCode = 1
	msg.Message.Data = []byte{}
	msg.Flag = 0
	//msg.Docs = []interface{}{}
	var args = "*"
	if req.FullCollectionName == "admin.$cmd" {
		msg.Docs = []interface{}{engine.command("admin", cmd, args)}
	}

	fields := strings.SplitN(req.FullCollectionName, ".", 2)

	_, cname := fields[0], fields[1]
	if strings.HasPrefix(cname, "system.") {
		msg.Docs = []interface{}{engine.command("system", cmd, args)}
	}
	if cname == "$cmd" {
		msg.Docs = []interface{}{engine.command("user", cmd, args)}
	}

	msg.NumberReturned = uint32(len(msg.Docs))
	return msg
}

func process2013(msg *MongoMessage) *M2013 {
	reader := bytes.NewReader(msg.Data)
	flagBits := readUInt32(reader)
	Debug("flagbit %d \r\n", flagBits)
	Meta, Sections := parseSection(reader)
	//Message := &msg
	// fix
	msg.Data = []byte{}
	return &M2013{Message: msg, FlagBits: int(flagBits),Meta: Meta, Sections: Sections }
}

func process2004 (msg *MongoMessage) *M2004 {
	reader := bytes.NewReader(msg.Data)
	flag := readUInt32(reader)
	Debug("flag %d \r\n", flag)
	FullCollectionName := readString(reader)
	Debug("colname %s \r\n", FullCollectionName)
	NumberToSkip := readUInt32(reader)
	NumberToReturn := readUInt32(reader)
	//Doc := make(bson.D)
	Debug("NumberToSkip %d \r\n", NumberToSkip)
	Debug("NumberToReturn %d \r\n", NumberToReturn)
	rs := &M2004{
		Message: msg,
		Flag: int(flag),
		FullCollectionName: FullCollectionName,
		NumberToSkip: NumberToSkip,
		NumberToReturn: NumberToReturn,
	}
	unmarshalDoc(reader, &rs.Doc)
	Debug("section %v \r\n", rs.Doc)
	if reader.Len() > 4 {
		unmarshalDoc(reader, &rs.ReturnFieldsSelector)
	}
	return rs
}

func write2013(w io.Writer, msg *M2013Reply) error {
	var out bytes.Buffer
	//https://docs.mongodb.com/manual/reference/mongodb-wire-protocol/#wire-msg-flags
	writeUint32(&out, uint32(msg.FlagBits)) //TODO FLAGBITS
	//out.Write([]byte{0, 0, 0, 0})
	Debug("write flag %d\r\n", msg.FlagBits)
	for i := 0; i < len(msg.sections); i++ {
		section := msg.sections[i]
		out.Write([]byte{0})
		bdata, _ := bson.Marshal(section)
		out.Write(bdata)
	}
	//out.Write([]byte{0,0,0,0})

	msg.Message.Data = out.Bytes()
	msg.Message.Length = uint32(out.Len() + 16)
	if needCheckSum(msg) { // bit match
		msg.Message.Length += 4
	}

	//var err error
	var out2 bytes.Buffer
	writeUint32(&out2, msg.Message.Length)
	writeUint32(&out2, msg.Message.RequestID)
	writeUint32(&out2, msg.Message.ResponseTo)
	writeUint32(&out2, msg.Message.OpCode)
	Debug("request_id %d response %d \r\n", msg.Message.RequestID, msg.Message.ResponseTo)
	out2.Write(msg.Message.Data)
	data2 := out2.Bytes()
	w.Write(data2)
	if needCheckSum(msg) { // bit match
		cs := getCheckSum(data2)
		w.Write(cs)
	}
	return nil
}

func write2001(w io.Writer, m *M2004Reply){
	var out bytes.Buffer
	writeUint32(&out, m.Flag)
	writeUint64(&out, m.CursorID)
	writeUint32(&out, m.StartingFrom)
	writeUint32(&out, m.NumberReturned)

	for _, doc := range m.Docs {
		b, err := bson.Marshal(doc)

		if err != nil {
			//return err
		}
		_, err = out.Write(b)
		if err != nil {
			//return err
		}
	}
	m.Message.Data = out.Bytes()
	m.Message.Length = uint32(out.Len() + 16)


	var out2 bytes.Buffer
	writeUint32(&out2, m.Message.Length)
	writeUint32(&out2, m.Message.RequestID)
	writeUint32(&out2, m.Message.ResponseTo)
	writeUint32(&out2, m.Message.OpCode)
	Debug("2001 request_id %d response %d \r\n", m.Message.RequestID, m.Message.ResponseTo)
	out2.Write(m.Message.Data)
	data2 := out2.Bytes()
	w.Write(data2)
}

func parse(c io.Reader, w io.Writer, engine MongoEngine) error {
	msg, err := readHeaderData(c)
	if err != nil {
		return err
	}
	Debug("metadata len[%d] from[%d]  to[%d] code[%d] \r\n", msg.Length, msg.RequestID, msg.ResponseTo, msg.OpCode)
	switch msg.OpCode {
	case 2013:
		req := process2013(msg)
		reply := handle2013(req, engine)
		write2013(w, reply)
	case 2004:
		req := process2004(msg)
		reply := handle2004(req, engine)
		write2001(w, reply)
	}
	return nil
}

func getCheckSum(data []byte) []byte {
	chesum := crc32.Checksum(data, crc32.MakeTable(crc32.Castagnoli))
	Debug("checksum %d \r\n", chesum)
	ca := make([]byte, 4)
	binary.LittleEndian.PutUint32(ca, chesum)
	return ca
}

func needCheckSum(msg *M2013Reply) bool {
	return msg.FlagBits != 0
}

func isPatternMatch(doc, pattern bson.M) bool {
	for matchKey, matchValue := range pattern {
		value, ok := doc[matchKey]
		if !ok || matchValue != value {
			return false
		}
	}
	return true
}


func updateDoc(doc bson.M ,modify bson.M) error{
	for k, v := range modify {
		unsuppErr := fmt.Errorf("unsupported update operator: %q", k)
		switch k {
		case "$currentDate":
			return unsuppErr
		case "$inc":
			return unsuppErr
		case "$max":
			return unsuppErr
		case "$min":
			return unsuppErr
		case "$mul":
			return unsuppErr
		case "$rename":
			return unsuppErr
		case "$setOnInsert":
			return unsuppErr
		case "$set":
			set, err := asBsonM(v)
			if err != nil {
				return err
			}
			for setK, setV := range set {
				doc[setK] = setV
			}
		case "$unset":
			return unsuppErr
		}
	}
	return fmt.Errorf("unsupported")
}