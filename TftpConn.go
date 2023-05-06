package main

import (
	"CSC445_Assignment2/tftp"
	"net"
	"time"
)

type TFTPConn struct {
	dhke       *DHKESession
	raddr      *net.UDPAddr
	pcktContet tftp.TFTPOpcode
	tOuts      int
	mDelay     time.Duration
	iDelay     time.Duration
	base       int
	nextSeqNum int
	curDelay   time.Duration
	connState  int
}

func NewTFTPConn(raddr *net.UDPAddr) *TFTPConn {
	return &TFTPConn{raddr: raddr}
}

func (tc *TFTPConn) InitializeBackOff() {
	tc.tOuts = 0
	tc.mDelay = 30 * time.Second
	tc.iDelay = 1 * time.Second
	tc.curDelay = tc.iDelay
}

func (tc *TFTPConn) SetBaseWindow() {
	tc.base = 1
	tc.nextSeqNum = 1
}

func (tc *TFTPConn) IncrSeqNum() {
	tc.nextSeqNum++
}

func (tc *TFTPConn) IncrBase() {
	tc.base++
}

func (tc *TFTPConn) IncrTouts() {
	tc.tOuts++
}

func (tc *TFTPConn) CalcExp() {
	tc.curDelay = tc.curDelay * 2
	if tc.curDelay > tc.mDelay {
		tc.curDelay = tc.mDelay
	}
}
