package go_modbus

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/VedrLabs/go_modbus/statute"
)

var NoConnectionError = errors.New("no connected")

// StatuteType 协议类型
type StatuteType string

const (
	ModbusTCP StatuteType = "modbusTCP"
	ModbusRTU StatuteType = "modbusRTU"
)

const (
	defaultRwTimeout      = 20 * time.Millisecond
	defaultConnectTimeout = 3 * time.Second
	defaultReadTimeout    = 2 * time.Second
	defaultWriteTimeout   = 2 * time.Second
)

// NewModbusTCPPacket 创建一个TCP连接
func NewModbusTCPPacket(ip string, port int, connectTimeout, readTimeout, writeTimeout, rwInterval time.Duration, modbusType StatuteType) (*ModbusTCPPacket, error) {
	if connectTimeout <= 0 {
		connectTimeout = defaultConnectTimeout
	}
	if readTimeout <= 0 {
		readTimeout = defaultReadTimeout
	}
	if writeTimeout <= 0 {
		writeTimeout = defaultWriteTimeout
	}
	if rwInterval <= 0 {
		rwInterval = defaultRwTimeout
	}
	tc := &ModbusTCPPacket{ip: ip, port: port, connectTimeout: connectTimeout, readTimeout: readTimeout, writeTimeout: writeTimeout, ModbusPacket: &ModbusPacket{rwInterval: rwInterval}}
	switch modbusType {
	case ModbusTCP:
		tc.ModbusCodec = statute.NewModbusTCPCodec()
	case ModbusRTU:
		tc.ModbusCodec = statute.NewModbusRTUCodec()
	default:
		return nil, errors.New("modbus type not supported")
	}
	tc.ModbusPacket.read = tc.read
	tc.ModbusPacket.write = tc.write
	return tc, nil
}

// ModbusTCPPacket MODBUS TCP
type ModbusTCPPacket struct {
	*ModbusPacket
	ip             string
	port           int
	connectTimeout time.Duration //连接超时
	readTimeout    time.Duration //读超时
	writeTimeout   time.Duration //写超时
	conn           *net.TCPConn
	reader         *bufio.Reader
}

func (T *ModbusTCPPacket) Connect() error {
	T.lock.Lock()
	defer T.lock.Unlock()
	conn, err := net.DialTimeout("tcp", T.ip+":"+strconv.Itoa(T.port), T.connectTimeout)
	if err != nil {
		return err
	}
	T.conn = conn.(*net.TCPConn)
	T.reader = bufio.NewReader(conn)
	return nil
}

func (T *ModbusTCPPacket) Close() error {
	T.lock.Lock()
	defer func() {
		T.reader = nil
		T.conn = nil
		T.lock.Unlock()
	}()
	if T.conn != nil {
		return T.conn.Close()
	}
	return nil
}

func (T *ModbusTCPPacket) write(frame []byte) (int, error) {
	if T.conn == nil {
		return 0, NoConnectionError
	}
	err := T.conn.SetWriteDeadline(time.Now().Add(T.writeTimeout))
	if err != nil {
		return 0, err
	}
	return T.conn.Write(frame)
}

func (T *ModbusTCPPacket) read() ([]byte, error) {
	if T.conn == nil {
		return nil, NoConnectionError
	}
	err := T.conn.SetReadDeadline(time.Now().Add(T.readTimeout))
	if err != nil {
		return nil, err
	}
	return T.ModbusCodec.Decode(T.reader)
}

func (T *ModbusTCPPacket) Flush() error {
	T.lock.Lock()
	defer T.lock.Unlock()
	if T.conn == nil {
		return NoConnectionError
	}
	_, err := io.ReadAll(T.reader)
	return err
}
