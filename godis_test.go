package godis

import (
    "testing"
    "bytes"
    "bufio"
    "os"
    "time"
    "fmt"
)

type simpleParserTest struct {
    in   string
    out  interface{}
    name string
    err  os.Error
}

func dummyReadWriter(data string) *redisReadWriter {
    r := bufio.NewReader(bytes.NewBufferString(data))
    w := bufio.NewWriter(bytes.NewBufferString(data))
    return &redisReadWriter{w, r}
}

var simpleParserTests = []simpleParserTest{
    {"+OK\r\n", "OK", "ok", nil},
    {"-ERR\r\n", nil, "err", os.NewError("ERR")},
    {":1\r\n", int64(1), "num", nil},
    {"$3\r\nfoo\r\n", s2Bytes("foo"), "bulk", nil},
    {"$-1\r\n", nil, "bulk-nil", nil},
    {"*-1\r\n", nil, "multi-bulk-nil", nil},
}

func TestParser(t *testing.T) {
    for _, test := range simpleParserTests {
        rw := dummyReadWriter(test.in)
        res, err := rw.read()

        if err != nil && test.err == nil {
            t.Errorf("'%s': unexpected error %v", test.name, err)
            t.FailNow()
        }

        switch v := res.(type) {
            case []byte:
                for i, c := range res.([]byte) {
                    if c != test.out.([]byte)[i] {
                        t.Errorf("expected %v got %v", test.out, res)
                    }
                }
            case [][]byte:
                for _, b := range res.([][]byte) {
                    for j, c := range b {
                        if c != test.out.([]byte)[j] {
                            t.Errorf("expected %v got %v", test.out, res)
                        }
                    }
                }
            default:
                if res != test.out {
                    t.Errorf("'%s': expected %s got %v", test.name, test.out, res)
                }
        }
        t.Log(test.in, res, test.out)
    }
}

func s2Bytes(s string) []byte {
    return bytes.NewBufferString(s).Bytes()
}

func s2MultiBytes(ss ...string) [][]byte {
    var buf = make([][]byte, len(ss))
    for i := 0; i < len(ss); i++ {
        buf[i] = s2Bytes(ss[i])
    }
    return buf
}

type SimpleSendTest struct {
    cmd  string
    args []string
    out  interface{}
}

var simpleSendTests = []SimpleSendTest{
    {"FLUSHDB", []string{}, "OK"},
    {"SET", []string{"key", "foo"}, "OK"},
    {"EXISTS", []string{"key"}, int64(1)},
    {"GET", []string{"key"}, s2Bytes("foo")},
    {"RPUSH", []string{"list", "foo"}, int64(1)},
    {"RPUSH", []string{"list", "bar"}, int64(2)},
    {"LRANGE", []string{"list", "0", "2"}, s2MultiBytes("foo", "bar")},
    {"KEYS", []string{"list"}, s2MultiBytes("list")},
    {"GET", []string{"/dev/null"}, nil},
}

func TestSimpleSend(t *testing.T) {
    c := New("", 0, "")
    for _, test := range simpleSendTests {
        res, err := c.Send(test.cmd, strToFaces(test.args)...)

        if err != nil {
            t.Errorf("'%s': unexpeced error %q", test.cmd, err)
            t.FailNow()
        }

        switch v := res.(type) {
        case []byte:
            for i, c := range res.([]byte) {
                if c != test.out.([]byte)[i] {
                    t.Errorf("'%s': expected %v got %v", test.cmd, test.out, res)
                }
            }
        case [][]byte:
            res_arr := res.([][]byte)
            out_arr := test.out.([][]byte)

            for i := 0; i < len(res_arr); i++ {
                for j := 0; j < len(res_arr[i]); j++ {
                    if res_arr[i][j] != out_arr[i][j] {
                        t.Errorf("'%s': expected %v got %v", test.cmd, test.out, res)
                    }
                }
            }
        default:
            if res != test.out {
                t.Errorf("'%s': expected %v got %v", test.cmd, test.out, res)
            }
        }
        t.Log(test.cmd, test.args, test.out)
    }
}

func BenchmarkParsing(b *testing.B) {
    c := New("", 0, "")

    for i := 0; i < 1000; i++ {
        c.Send("RPUSH", "list", "foo")
    }

    start := time.Nanoseconds()

    for i := 0; i < b.N; i++ {
        c.Send("LRANGE", "list", "0", "50")
    }

    stop := time.Nanoseconds() - start

    fmt.Fprintf(os.Stdout, "time: %.2f\n", float32(stop / 1.0e+6) / 1000.0)
    c.Send("FLUSHDB")
}


