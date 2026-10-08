package go_modbus

import (
	"errors"
	"sync"
	"time"

	"github.com/VedrLabs/go_modbus/statute"
)

type ModbusPacket struct {
	lock sync.Mutex
	statute.ModbusCodec
	write      func([]byte) (int, error)       //写
	rwInterval time.Duration                   //读写间隔
	read       func() (data []byte, err error) //读
}

// 读写
func (T *ModbusPacket) wr(frame []byte) (data []byte, err error) {
	T.lock.Lock()
	defer T.lock.Unlock()
	if _, err = T.write(frame); err != nil {
		return nil, err
	}
	time.Sleep(T.rwInterval)
	return T.read()
}

// ReadCoils 读线圈
// slaveId 从站id
// address 寄存器起始地址
// number 寄存器数量
func (T *ModbusPacket) ReadCoils(slaveId byte, address, number uint16) (length uint16, result []statute.CoilStatus, err error) {
	req := T.BuildReadCoils(slaveId, address, number)
	data, err := T.wr(req)
	if err != nil {
		return 0, nil, err
	}
	return T.ObtainIntermediary().ParseReadCoilsResponse(number, data)
}

// ReadDiscreteInputs 读离散输入寄存器
// slaveId 从站id
// address 寄存器起始地址
// number 寄存器数量
func (T *ModbusPacket) ReadDiscreteInputs(slaveId byte, address, number uint16) (length uint16, result []statute.CoilStatus, err error) {
	req := T.BuildReadDiscreteInputs(slaveId, address, number)
	data, err := T.wr(req)
	if err != nil {
		return 0, nil, err
	}
	return T.ObtainIntermediary().ParseReadDiscreteInputsResponse(number, data)
}

// ReadHoldingRegisters 读保持寄存器
// slaveId 从站id
// addr 寄存器起始地址
// number 寄存器数量
func (T *ModbusPacket) ReadHoldingRegisters(slaveId byte, address, number uint16) ([]byte, error) {
	req := T.BuildReadHoldingRegisters(slaveId, address, number)
	data, err := T.wr(req)
	if err != nil {
		return nil, err
	}
	if uint16(len(data)) != number*2 {
		return nil, errors.New("ReadHoldingRegisters: length mismatch")
	}
	return data, nil
}

// ReadInputRegisters 读输入寄存器
// slaveId 从站id
// addr 寄存器起始地址
// number 寄存器数量
func (T *ModbusPacket) ReadInputRegisters(slaveId byte, address, number uint16) ([]byte, error) {
	req := T.BuildReadInputRegisters(slaveId, address, number)
	data, err := T.wr(req)
	if err != nil {
		return nil, err
	}
	if uint16(len(data)) != number*2 {
		return nil, errors.New("ReadHoldingRegisters: length mismatch")
	}
	return data, nil
}

// WriteSingleCoil 写单个线圈
// slaveId 从站id
// addr 起始地址
// value 设定值 写0xFF00表示线圈为ON，写0x0000表示线圈为OFF
// 返回值
// addr 地址
// status true-ON false-OFF
func (T *ModbusPacket) WriteSingleCoil(slaveId byte, address uint16, value statute.CoilStatus) (addr uint16, status statute.CoilStatus, err error) {
	req := T.BuildWriteSingleCoil(slaveId, address, value)
	data, err := T.wr(req)
	if err != nil {
		return 0, statute.OFF, err
	}
	addrValue, statusValue, err := T.ObtainIntermediary().ParseWriteSingleCoil(data)
	if err != nil {
		return 0, statute.OFF, err
	}
	if statusValue == 0xFF00 {
		status = statute.ON
	} else if statusValue == 0x0000 {
		status = statute.OFF
	} else {
		return 0, false, errors.New("invalid status")
	}
	return addrValue, status, nil
}

// WriteSingleRegister 写单个保持寄存器
// slaveId 从站id
// addr 寄存器起始地址
// value 设定值
func (T *ModbusPacket) WriteSingleRegister(slaveId byte, address uint16, value uint16) (addr, status uint16, err error) {
	req := T.BuildWriteSingleRegister(slaveId, address, value)
	data, err := T.wr(req)
	if err != nil {
		return 0, 0, err
	}
	return T.ObtainIntermediary().ParseWriteSingleRegister(data)
}

// WriteMultipleCoils 写多个线圈的请求
// slaveId 从站id
// addr 寄存器起始地址
// status 线圈状态
func (T *ModbusPacket) WriteMultipleCoils(slaveId byte, address uint16, status ...statute.CoilStatus) (addr, size uint16, err error) {
	req, err := T.BuildWriteMultipleCoils(slaveId, address, status...)
	if err != nil {
		return 0, 0, err
	}
	data, err := T.wr(req)
	if err != nil {
		return 0, 0, err
	}
	return T.ObtainIntermediary().ParseWriteMultipleCoilsResponse(data)
}

// WriteMultipleRegisters 写多个保持寄存器
func (T *ModbusPacket) WriteMultipleRegisters(slaveId byte, address uint16, value ...uint16) (addr, number uint16, err error) {
	req, err := T.BuildWriteMultipleRegisters(slaveId, address, value...)
	if err != nil {
		return 0, 0, err
	}
	resp, err := T.wr(req)
	if err != nil {
		return 0, 0, err
	}
	return T.ObtainIntermediary().PraseWriteMultipleRegisters(resp)
}
