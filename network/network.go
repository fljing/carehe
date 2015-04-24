// network project network.go
package network

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"sync"
)

const (
	BufferSize = 4096
	PacketSize = 1 * 1024 * 1024
)

type CareheServer struct {
	listener net.Listener
	chexit   chan bool
	wg       sync.WaitGroup
	network  string
	laddr    string
	run      bool
	handler  CareheHandler
}

type CareheClient struct {
	multi   *Connection
	handler CareheHandler
}

type CareheHandler interface {
	Process(*CareheClient, []byte)
	OnError(*CareheClient, error)
	Connect(*CareheClient)
	Disconnect(*CareheClient)
}

func (p CareheClient) Process(data []byte) {
	p.handler.Process(&p, data)
}

func (p CareheClient) OnError(err error) {
	p.handler.OnError(&p, err)
}

type Client struct {
	conn     net.Conn
	identity map[string]uint64
}

func (p *Client) Read() ([]byte, error) {
	msgbuf := bytes.NewBuffer(make([]byte, 0, PacketSize))
	length := uint32(0)
	lenbuf := make([]byte, 4)

	for {
		if length == 0 {
			n, err := p.conn.Read(lenbuf)
			if err != nil {
				return nil, err
			}
			n, err = msgbuf.Write(lenbuf[:n])
			if msgbuf.Len() >= 4 {
				binary.Read(msgbuf, binary.LittleEndian, &length)
				if length < 0 || length > PacketSize {
					return nil, errors.New("Packet Size Error")
				}
			}
		} else {
			databuf := make([]byte, length)
			n, err := p.conn.Read(databuf)
			if err != nil {
				return nil, err
			}

			n, err = msgbuf.Write(databuf[:n])
			if err != nil {
				fmt.Printf("Buffer write error: %s\n", err.Error())
				return nil, err
			}

			if length > 0 && uint32(msgbuf.Len()) >= length {
				msg := make([]byte, length)
				rn := 0
				rn, err = msgbuf.Read(msg)
				if err != nil {
					fmt.Printf("Copy data error: %s\n", err.Error())
					return nil, err
				} else {
					if uint32(rn) != length {
						fmt.Printf("Copy data length error: copy length=%d, recv length=%d\n", rn, length)
						return nil, errors.New("Copy data error")
					}
				}
				//fmt.Printf("%d\n", length)
				//go this.HandlerMsg(cmd, msg, conn)
				return msg, nil
			}
		}
	}
}

func (p *Client) Write(data []byte) error {
	length := uint32(len(data))
	if length > PacketSize {
		return errors.New("Write data length too long")
	}
	sendLength := length + 4
	msgbuf := bytes.NewBuffer(make([]byte, 0, sendLength))
	binary.Write(msgbuf, binary.LittleEndian, length)
	msgbuf.Write(data)
	sendData := msgbuf.Bytes()
	var sn uint32 = 0
	for {
		if sn < sendLength {
			n, err := p.conn.Write(sendData[sn:])
			if err != nil {
				return err
			}
			sn += uint32(n)
		}
		if sn == sendLength {
			break
		}
		if sn > sendLength {
			return errors.New("already send length more than need send length")
		}
	}
	return nil
}

func (p *Client) Close() {
	p.conn.Close()
}

func (p *Client) GetIdentity(data []byte) uint64 {
	m := md5.New()
	m.Write(data)
	return p.identity[hex.EncodeToString(m.Sum(nil))]
}

func (p *Client) SetIdentity(data []byte, id uint64) {

	m := md5.New()
	m.Write(data)
	p.identity[hex.EncodeToString(m.Sum(nil))] = id
}

func NewCareheServer(network, laddr string, handler CareheHandler) *CareheServer {
	return &CareheServer{network: network, laddr: laddr, chexit: make(chan bool), handler: handler}
}

func NewCareheClient(handler CareheHandler) *CareheClient {
	return &CareheClient{handler: handler}
}

func (p *CareheClient) Connect(network, address string) error {
	var err error
	conn, err := net.Dial(network, address)
	if err != nil {
		return err
	}
	client := &Client{conn: conn, identity: make(map[string]uint64)}
	p.multi = NewConnection(client, 1024, p, client, p)
	p.multi.Start()
	return nil
}

func (p *CareheServer) Start() {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.Listen()
	}()
}

func (p *CareheServer) Listen() error {
	var err error = nil
	p.listener, err = net.Listen(p.network, p.laddr)
	if err != nil {
		return err
	}
	p.run = true
	defer func() {
		p.run = false
	}()

	for {
		select {
		case <-p.chexit:
			{
				return errors.New("exit")
			}
		default:
			break
		}
		conn, err := p.listener.Accept()
		if err != nil {
			fmt.Printf("Accept error: %s\n", err.Error())
			continue
		}
		careheClient := NewCareheClient(p.handler)
		client := &Client{conn: conn, identity: make(map[string]uint64)}
		careheClient.multi = NewConnection(client, 1024, careheClient, client, careheClient)
		careheClient.handler.Connect(careheClient)
		careheClient.Start()
	}
	return nil
}

func (p *CareheServer) Stop() {
	close(p.chexit)
	p.listener.Close()
	p.wg.Wait()
}

func (p *CareheServer) IsRun() bool {
	return p.run
}

func (p *CareheClient) Start() {
	p.multi.Start()
}

func (p *CareheClient) Close() {
	p.multi.Close()
}

func (p *CareheClient) Query(data []byte) (res []byte, err error) {
	res, err = p.multi.Query(data)
	return
}

func (p *CareheClient) Write(data []byte) error {
	return p.multi.Write(data)
}

func (p *CareheClient) Reply(query, answer []byte) error {
	return p.multi.Reply(query, answer)
}
