package ngap

import (
	"encoding/json"
	"io/ioutil"
	"math/bits"
	"fmt"
	"log"
	"strconv"
	"../encoding/per"
)

const (
	reject = iota
	ignore
	notify
)

const (
	initiatingMessage = iota
	sucessfulOutcome
	unsuccessfulOutcome
)

// Elementary Procedures constants
const (
	procCodeInitialUEMessage = 15
	procCodeNGSetup          = 21
)

const (
	idDefaultPagingDRX = 21
	idGlobalRANNodeID  = 27
    idSupportedTAList  = 102
)

const (
	globalGNB = iota
	globalNgGNB
	globalN3IWF
)

type GNB struct {
	GlobalGNBID GlobalGNBID
	SupportedTAList []SupportedTA
}

type GlobalGNBID struct {
	MCC uint16
	MNC uint16
	GNBID uint32
}

type SupportedTA struct {
	TAC string
	BroadcastPLMNList []BroadcastPLMN
}

type BroadcastPLMN struct {
	MCC uint16
	MNC uint16
	SliceSupportList []SliceSupport
}

type SliceSupport struct {
	SST uint8
	SD string
}

func InitNGAP(filename string) (p *GNB) {

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	var gnb GNB
	p = &gnb
	json.Unmarshal(bytes, p)

	return
}

/*
NGAP-PDU ::= CHOICE {
    initiatingMessage           InitiatingMessage,
    successfulOutcome           SuccessfulOutcome,
    unsuccessfulOutcome         UnsuccessfulOutcome,
    ...
}
ProcedureCode ::= INTEGER (0..255)
Criticality   ::= ENUMERATED { reject, ignore, notify }

InitiatingMessage ::= SEQUENCE {
    procedureCode   NGAP-ELEMENTARY-PROCEDURE.&procedureCode        ({NGAP-ELEMENTARY-PROCEDURES}),
    criticality     NGAP-ELEMENTARY-PROCEDURE.&criticality          ({NGAP-ELEMENTARY-PROCEDURES}{@procedureCode}),
    value           NGAP-ELEMENTARY-PROCEDURE.&InitiatingMessage    ({NGAP-ELEMENTARY-PROCEDURES}{@procedureCode})
}
*/
func encNgapPdu(pduType, procCode, criticality int) (pdu []uint8) {
	pdu, _, _ = per.EncChoice(pduType, 0, 2, true)
	v, _, _ := per.EncInteger(procCode, 0, 255, false)
	pdu = append(pdu, v...)
	v, _, _ = per.EncEnumerated(criticality, 0, 2, false)
	pdu = append(pdu, v...)

	return
}

/*
NGSetupRequest ::= SEQUENCE {
    protocolIEs     ProtocolIE-Container        { {NGSetupRequestIEs} },
    ...
}
ProtocolIE-Container {NGAP-PROTOCOL-IES : IEsSetParam} ::= 
    SEQUENCE (SIZE (0..maxProtocolIEs)) OF
    ProtocolIE-Field {{IEsSetParam}}

maxProtocolIEs                          INTEGER ::= 65535
*/
func encProtocolIEContainer(num int) (container []uint8) {
	const maxProtocolIEs = 65535
	container, _, _ = per.EncSequence(true, 0, 0)
	v, _, _ := per.EncSequenceOf(num, 0, maxProtocolIEs, false)
	container = append(container, v...)

	return
}

// 9.2.6.1 NG SETUP REQUEST
/*
NGSetupRequestIEs NGAP-PROTOCOL-IES ::= {
    { ID id-GlobalRANNodeID         CRITICALITY reject  TYPE GlobalRANNodeID            PRESENCE mandatory  }|
    { ID id-RANNodeName             CRITICALITY ignore  TYPE RANNodeName                PRESENCE optional}|
    { ID id-SupportedTAList         CRITICALITY reject  TYPE SupportedTAList            PRESENCE mandatory  }|
    { ID id-DefaultPagingDRX        CRITICALITY ignore  TYPE PagingDRX                  PRESENCE mandatory  }|
    { ID id-UERetentionInformation  CRITICALITY ignore  TYPE UERetentionInformation     PRESENCE optional   },
    ...
}
*/
func MakeNGSetupRequest(p *GNB) {

	pdu := encNgapPdu(initiatingMessage, procCodeNGSetup, reject)
	fmt.Printf("result: pdu = %02x\n", pdu)

	v := encProtocolIEContainer(3)
	fmt.Printf("result: ie container = %02x\n", v)

	v, _ = encGlobalRANNodeID(&p.GlobalGNBID)
	fmt.Printf("result: global RAN Node ID = %02x\n", v)

	v, _ = encSupportedTAList(&p.SupportedTAList)
	fmt.Printf("result: Supported TA List = %02x\n", v)
}

/*
BroadcastPLMNList ::= SEQUENCE (SIZE(1..maxnoofBPLMNs)) OF BroadcastPLMNItem
    maxnoofBPLMNs                       INTEGER ::= 12
 */
func encBroadcastPLMNList(p *[]BroadcastPLMN) (v []uint8) {
	const maxnoofBPLMNs = 12
	v, _, _ = per.EncSequenceOf(1, 1, maxnoofBPLMNs, false)

	for _, item := range *p {
		v = append(v, encBroadcastPLMNItem(&item)...)
	}
	return
}

/*
BroadcastPLMNItem ::= SEQUENCE {
    pLMNIdentity            PLMNIdentity,
    tAISliceSupportList     SliceSupportList,
    iE-Extensions           ProtocolExtensionContainer { {BroadcastPLMNItem-ExtIEs} } OPTIONAL,
    ...
}
 */
func encBroadcastPLMNItem(p *BroadcastPLMN) (v []uint8) {
	v, _, _ = per.EncSequence(true, 1, 0)
	v = append(v, encPLMNIdentity(p.MCC, p.MNC)...)
	v = append(v, encSliceSupportList(&p.SliceSupportList)...)
	return
}

func encProtocolIE(id, criticality int) (v []uint8, err error) {

	v1, _, _ := per.EncInteger(id, 0, 65535, false)
	v2, _, _ := per.EncEnumerated(criticality, 0, 2, false)
	v = append(v1, v2...)

	return
}

// 9.3.1.5 Global RAN Node ID
/*
  It returns only GNB-ID for now.
   GlobalRANNodeID ::= CHOICE {
       globalGNB-ID        GlobalGNB-ID,
       globalNgENB-ID      GlobalNgENB-ID,
       globalN3IWF-ID      GlobalN3IWF-ID,
       choice-Extensions   ProtocolIE-SingleContainer { {GlobalRANNodeID-ExtIEs} }
   }
 */
func encGlobalRANNodeID(p *GlobalGNBID) (v []uint8, err error) {

	v, err = encProtocolIE(idGlobalRANNodeID, reject)

	// NG-ENB and N3IWF are not implemented yet...
	pv, plen, _ := per.EncChoice(globalGNB, 0, 2, false)
	pv2, plen2, v2 := encGlobalGNBID(p)
	pv, plen = per.MergeBitField(pv, plen, pv2, plen2)
	pv = append(pv, v2...)

	v3, _, _ := per.EncLengthDeterminant(len(pv), 0)
	v = append(v, v3...)
	v = append(v, pv...)

	return
}


// 9.3.1.6 Global gNB ID
/*
   GlobalGNB-ID ::= SEQUENCE {
       pLMNIdentity        PLMNIdentity,
       gNB-ID              GNB-ID,
       iE-Extensions       ProtocolExtensionContainer { {GlobalGNB-ID-ExtIEs} } OPTIONAL,
       ...
   }
 */
func encGlobalGNBID(p *GlobalGNBID) (pv []uint8, plen int, v []uint8) {

	pv, plen, _ = per.EncSequence(true, 1, 0)
	v = append(v, encPLMNIdentity(p.MCC, p.MNC)...)

	pv2, _ := encGNBID(p.GNBID)
	v = append(v, pv2...)
	return
}

/*
   GNB-ID ::= CHOICE {
       gNB-ID                  BIT STRING (SIZE(22..32)),
       choice-Extensions       ProtocolIE-SingleContainer { {GNB-ID-ExtIEs} }
   }
 */
func encGNBID(gnbid uint32) (pv []uint8, plen int) {
	const minGNBIDSize = 22
	const maxGNBIDSize = 22

	bitlen := bits.Len32(gnbid)
	if bitlen < minGNBIDSize {
		bitlen = minGNBIDSize
	}

	pv, plen, _ = per.EncChoice(0, 0, 1, false)

	tmp := per.IntToByte(gnbid)
	pv2, plen2, _ := per.EncBitString(tmp, bitlen, 22, 32, false)

	pv, plen = per.MergeBitField(pv, plen, pv2, plen2)

	return
}

// 9.3.1.90 PagingDRX
/*
func encPagingDRX(drx string) (val []uint8) {
	n := 0
	switch drx {
	case "v32":
		n = 0
	case "v64":
		n = 1
	case "v128":
		n = 2
	case "v256":
		n = 3
	default:
		fmt.Printf("encPagingDRX: no such DRX value(%s)", drx)
		return
	}
	val = per.EncEnumerated(n)
	return
}
*/

// 9.3.3.5 PLMN Identity
/*
PLMNIdentity ::= OCTET STRING (SIZE(3)) 
 */
func encPLMNIdentity(mcc, mnc uint16) (v []uint8) {

	v = make([]uint8, 3, 3)
	v[0] = uint8(mcc % 1000 / 100)
	v[0] |= uint8(mcc%100/10) << 4

	v[1] = uint8(mcc % 10)
	v[1] |= 0xf0 // filler digit

	v[2] = uint8(mnc % 100 / 10)
	v[2] |= uint8(mnc%10) << 4

	_, _, v, _ = per.EncOctetString(v, 3, 3, false)

	return
}

/*
SliceSupportList ::= SEQUENCE (SIZE(1..maxnoofSliceItems)) OF SliceSupportItem
    maxnoofSliceItems                   INTEGER ::= 1024
 */
func encSliceSupportList(p *[]SliceSupport) (v []uint8) {
	v, _, _ = per.EncSequenceOf(1, 1, 1024, false)
	for _, item := range *p {
		v = append(v, encSliceSupportItem(&item)...)
	}
	return
}

/*
SliceSupportItem ::= SEQUENCE {
    s-NSSAI             S-NSSAI,
    iE-Extensions       ProtocolExtensionContainer { {SliceSupportItem-ExtIEs} }    OPTIONAL,
    ...
}
 */
func encSliceSupportItem(p *SliceSupport) (v []uint8) {
	/*
	ex.1
	    .    .   .          .    .   .    .   .    .   .
	00 000 00000001	00 010 00000010 00000000 00000011 00001000
	0000 0000 0000 1xxx
	                000 1000 0000 10xx xxxx 00000000 00000011 11101000
	0x00 0x08 0x80 0x80 0x00 0x03 0xe1

	ex.2
	    .    .   .    .        .        .        .
	00 010 00000001 xxx 00000000 00000000 01111011
	0001 0000 0000 1xxx 00000000 00000000 11101000
	0x10 0x08 0x80 0x00 0x00 0x7b
	*/
	pv, plen, _ := per.EncSequence(true, 1, 0)

	pv2, plen2, v := encSNSSAI(p.SST, p.SD)
	pv, plen = per.MergeBitField(pv, plen, pv2, plen2)
	v = append(pv, v...)
	return
}

// 9.3.1.24 S-NSSAI
/*
S-NSSAI ::= SEQUENCE {
    sST           SST,
    sD            SD                                                  OPTIONAL,
    iE-Extensions ProtocolExtensionContainer { { S-NSSAI-ExtIEs} }    OPTIONAL,
    ...
}

SST ::= OCTET STRING (SIZE(1))
SD ::= OCTET STRING (SIZE(3))
*/
func encSNSSAI(sstInt uint8, sdString string) (pv []uint8, plen int, v []uint8) {
	pv, plen, _ = per.EncSequence(true, 2, 0x02)

	sst := per.IntToByte(sstInt)
	pv2, plen2, _, _ := per.EncOctetString(sst, 1, 1, false)

	pv, plen = per.MergeBitField(pv, plen, pv2, plen2)

	tmp, _ := strconv.ParseUint(sdString, 0, 32)
	sd := per.IntToByte(tmp)
	_, _, v, _ = per.EncOctetString(sd, 3, 3, false)
	return
}

// Supported TA List
/*
SupportedTAList ::= SEQUENCE (SIZE(1..maxnoofTACs)) OF SupportedTAItem
 */
func encSupportedTAList(p *[]SupportedTA) (v []uint8, err error) {

	v, err = encProtocolIE(idSupportedTAList, reject)

	// maxnoofTACs INTEGER ::= 256
	const maxnoofTACs = 256
	pv, _, _ := per.EncSequenceOf(1, 1, maxnoofTACs, false)
	v = append(v, pv...)

	for _, item := range *p {
		v = append(v, encSupportedTAItem(&item)...)
	}

	return
}

// Supported TA Item
/*
SupportedTAItem ::= SEQUENCE {
    tAC                     TAC,
    broadcastPLMNList       BroadcastPLMNList,
    iE-Extensions           ProtocolExtensionContainer { {SupportedTAItem-ExtIEs} } OPTIONAL,
    ...
}
 */
func encSupportedTAItem(p *SupportedTA) (v []uint8) {

	pv, _, _ := per.EncSequence(true, 1, 0)
	v = append(pv, encTAC(p.TAC)...)
	v = append(v, encBroadcastPLMNList(&p.BroadcastPLMNList)...)
	return
}

// 9.3.3.10 TAC
/*
TAC ::= OCTET STRING (SIZE(3))
 */
func encTAC(tacString string) (v []uint8) {
	const tacSize = 3
	tmp, _ := strconv.ParseUint(tacString, 0, 32)
	tac := per.IntToByte(tmp)
	tac = tac[:1]

	_, _, v, _ = per.EncOctetString(tac, tacSize, tacSize, false)
	return
}

