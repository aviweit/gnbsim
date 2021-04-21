package main

import (
	"fmt"
	"github.com/hhorai/gnbsim/encoding/gtp"
	"github.com/hhorai/gnbsim/encoding/nas"
	"github.com/hhorai/gnbsim/encoding/ngap"
	"github.com/wmnsk/go-gtp/gtpv1"
	"github.com/ishidawataru/sctp"
	"log"
	"net"
	"time"
)

type testSession struct {
	conn *sctp.SCTPConn
	info *sctp.SndRcvInfo
	gnb  *ngap.GNB
	ue   *nas.UE
	gtpu *gtp.GTP
	uConn *gtpv1.UPlaneConn
}

func setupSCTP(gnb *ngap.GNB) (conn *sctp.SCTPConn, info *sctp.SndRcvInfo) {
	log.Printf("setupSCTP function called")
	log.Printf("AMF address is %s",gnb.NGAPPeerAddr)

	const amfPort = 38412
	amfAddr, _ := net.ResolveIPAddr("ip", gnb.NGAPPeerAddr)

	ips := []net.IPAddr{*amfAddr}
	addr := &sctp.SCTPAddr{
		IPAddrs: ips,
		Port:    amfPort,
	}

	conn, err := sctp.DialSCTP("sctp", nil, addr)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	log.Printf("Dail LocalAddr: %s; RemoteAddr: %s",
		conn.LocalAddr(), conn.RemoteAddr())

	ppid := 0
	info = &sctp.SndRcvInfo{
		Stream: uint16(ppid),
		PPID:   0x3c000000,
	}

	conn.SubscribeEvents(sctp.SCTP_EVENT_DATA_IO)

	return
}

func (t *testSession) sendtoAMF(pdu []byte) {
	log.Printf("sendtoAMF function called")

	n, err := t.conn.SCTPWrite(pdu, t.info)
	if err != nil {
		log.Fatalf("failed to write: %v", err)
	}
	log.Printf("write: len %d, info: %+v", n, t.info)
	return
}

func (t *testSession) recvfromAMF(timeout time.Duration) {
	log.Printf("recvfromAMF function called")

	const defaultTimer = 10 // sec

	if timeout == 0 {
		timeout = defaultTimer
	}

	c := make(chan bool, 1)
	go func() {
		buf := make([]byte, 1500)
		n, info, err := t.conn.SCTPRead(buf)
		t.info = info

		if err != nil {
			log.Fatalf("failed to read: %v", err)
		}
		log.Printf("read: len %d, info: %+v", n, t.info)

		buf = buf[:n]
		fmt.Printf("dump: %x\n", buf)
		t.gnb.Decode(&buf)
		c <- true
	}()
	select {
	case <-c:
		break
	case <-time.After(timeout * time.Second):
		log.Printf("read: timeout")
	}
	return
}

func initRAN() (t *testSession) {
	log.Printf("Init RAN function called")

	t = new(testSession)
	gnb := ngap.NewNGAP("gnb.json")
	log.Printf("read gnb.json")
	log.Printf("GNB Values \n%v\n",gnb)
	gnb.SetDebugLevel(1)

	conn, info := setupSCTP(gnb)

	t.gnb = gnb
	t.conn = conn
	t.info = info

	pdu := gnb.MakeNGSetupRequest()
	t.sendtoAMF(pdu)
	t.recvfromAMF(0)

	return
}

func (t *testSession) updateNGAP(filename string) {
	log.Printf("Calling UpdateNGAP")

	newgnb := ngap.NewNGAP(filename)

	t.gnb.UE=newgnb.UE
	t.gnb.GTPuLocalAddr=newgnb.GTPuLocalAddr

	log.Printf("Updating gnb to %v", t.gnb)

	return
}