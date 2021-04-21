package main

import (
	"log"
	"github.com/hhorai/gnbsim/encoding/nas"
)

func (t *testSession) registerUE(ue *nas.UE) {
	log.Printf("registerUE function called for %v",ue)

	log.Printf("send registration request -->")
	pdu := ue.MakeRegistrationRequest()
	t.gnb.RecvfromUE(ue,&pdu)

	log.Printf("send initial UE message -->")
	buf := t.gnb.MakeInitialUEMessage(ue)
	t.sendtoAMF(buf)
	log.Printf("receive initial UE message <--")
	t.recvfromAMF(0)

	log.Printf("receive authentication response <--")
	pdu = ue.MakeAuthenticationResponse()
	t.gnb.RecvfromUE(ue,&pdu)

	log.Printf("send uplink NAS transport -->")
	buf = t.gnb.MakeUplinkNASTransport(ue)
	t.sendtoAMF(buf)
	log.Printf("receive uplink NAS transport <--")
	t.recvfromAMF(0)

	log.Printf("receive security mode complete <--")
	pdu = ue.MakeSecurityModeComplete()
	t.gnb.RecvfromUE(ue,&pdu)

	log.Printf("send uplink NAS transport -->")
	buf = t.gnb.MakeUplinkNASTransport(ue)
	t.sendtoAMF(buf)
	log.Printf("receive uplink NAS transport <--")
	t.recvfromAMF(0)

	log.Printf("send initial context setup -->")
	buf = t.gnb.MakeInitialContextSetupResponse(ue)
	t.sendtoAMF(buf)

	log.Printf("receive registration complete <--")
	pdu = ue.MakeRegistrationComplete()
	t.gnb.RecvfromUE(ue,&pdu)

	log.Printf("send uplink NAS transport -->")
	buf = t.gnb.MakeUplinkNASTransport(ue)
	t.sendtoAMF(buf)

	log.Printf("receive configuration command <--")
	// for Configuration Update Command from open5gs AMF.
	t.recvfromAMF(3)

	return
}


func (t *testSession) establishPDUSession(ue *nas.UE) {
	log.Printf("establishPDUSession RAN function called")

	log.Printf("receive PDU session <--")
	pdu := ue.MakePDUSessionEstablishmentRequest()
	t.gnb.RecvfromUE(ue,&pdu)

	log.Printf("send uplink NAS transport -->")
	buf := t.gnb.MakeUplinkNASTransport(ue)
	t.sendtoAMF(buf)
	log.Printf("receive uplink ack <--")
	t.recvfromAMF(0)

	log.Printf("send PDU session setup -->")
	buf = t.gnb.MakePDUSessionResourceSetupResponse(ue)
	t.sendtoAMF(buf)

	return
}

func (t *testSession) initUE() {
	log.Printf("Init UE function called")
	gnb := t.gnb
	tmp := t.gnb.UE
	ue := &tmp
	ue.PowerON()
	ue.SetDebugLevel(1)
	gnb.CampIn(ue)

	return
}
