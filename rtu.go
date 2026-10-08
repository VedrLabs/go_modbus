package go_modbus

import (
	"bufio"
	"errors"
	"time"

	"github.com/VedrLabs/go_modbus/statute"
	"github.com/tarm/serial"
)

type Parity byte

// NewModbusRTUPacket 创建一个RTU连接
func NewModbusRTUPacket(port string, baud int, dataBit byte, parity Parity, stopBit byte, readTimeout, rwInterval time.Duration, modbusType StatuteType) (*ModbusRTUPacket, error) {
	if readTimeout <= 0 {
		readTimeout = defaultReadTimeout
	}
	if rwInterval <= 0 {
		rwInterval = defaultRwTimeout
	}
	tc := &ModbusRTUPacket{
		ModbusPacket: &ModbusPacket{rwInterval: rwInterval},
		port:         port,
		baud:         baud,
		dataBit:      dataBit,
		parity:       parity,
		stopBit:      stopBit,
		readTimeout:  readTimeout,
	}
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

type ModbusRTUPacket struct {
	*ModbusPacket

	port        string        //串口号
	baud        int           //波特率
	dataBit     byte          //数据位
	parity      Parity        //校验位
	stopBit     byte          //停止位
	readTimeout time.Duration //读超时

	serialPort *serial.Port

	reader *bufio.Reader
}

func (T *ModbusRTUPacket) Connect() error {
	config := &serial.Config{Name: T.port, Baud: T.baud, Size: T.dataBit, Parity: serial.Parity(T.parity), StopBits: serial.StopBits(T.stopBit), ReadTimeout: T.readTimeout}
	port, err := serial.OpenPort(config)
	if err != nil {
		return err
	}
	T.serialPort = port
	T.reader = bufio.NewReader(port)
	return nil
}

func (T *ModbusRTUPacket) Close() error {
	T.lock.Lock()
	defer func() {
		T.reader = nil
		T.serialPort = nil
		T.lock.Unlock()
	}()
	if T.serialPort != nil {
		return T.serialPort.Close()
	}
	return nil
}

func (T *ModbusRTUPacket) write(frame []byte) (int, error) {
	if T.serialPort == nil {
		return 0, NoConnectionError
	}
	return T.serialPort.Write(frame)
}

func (T *ModbusRTUPacket) read() ([]byte, error) {
	if T.serialPort == nil {
		return nil, NoConnectionError
	}
	return T.ModbusCodec.Decode(T.reader)
}

func (T *ModbusRTUPacket) Flush() error {
	T.lock.Lock()
	defer T.lock.Unlock()
	if T.serialPort == nil {
		return NoConnectionError
	}
	return T.serialPort.Flush()
}
