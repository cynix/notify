package main

import (
	"log"
	"net"

	"github.com/gosnmp/gosnmp"
	//"github.com/sleepinggenius2/gosmi"
)


type SnmpListener struct {
	l *gosnmp.TrapListener
	addr string
}

func NewSnmpListener(addr string) (Listener, error) {
	return &SnmpListener{gosnmp.NewTrapListener(), addr}, nil
}

func (s *SnmpListener) Listen(ch chan<- Message) error {
	s.l.OnNewTrap = s.trap
	return s.l.Listen(s.addr)
}

func (s *SnmpListener) Close() error {
	s.l.Close()
	return nil
}

func (s *SnmpListener) trap(p *gosnmp.SnmpPacket, u *net.UDPAddr) {
	for _, v := range p.Variables {
		log.Printf("snmp trap: name=%v type=%v value=%v", v.Name, v.Type, v.Value)
	}
}
