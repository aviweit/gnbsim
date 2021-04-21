package main

import (
	"context"
	"github.com/hhorai/gnbsim/encoding/gtp"
	"github.com/hhorai/gnbsim/encoding/ngap"
	"github.com/wmnsk/go-gtp/gtpv1"
	"github.com/vishvananda/netlink"
	"log"
	"net"
	"time"
)

func runUserPlane(t *testSession, gnbAddress string, gtpInterface string, ctx context.Context, c*ngap.Camper) error {
	log.Printf("runUserPlane function called with variables %s %s", gnbAddress, gtpInterface)

	gnb := t.gnb
	ue := c.UE

	localAddress := &net.UDPAddr{
		IP:   net.ParseIP(gnbAddress),
		Port: gtp.Port,
	}
	log.Printf("gNB UDP local address: %v\n", localAddress)

	conn := gtpv1.NewUPlaneConn(localAddress)
	if err := conn.EnableKernelGTP(gtpInterface, gtpv1.RoleSGSN); err != nil {
		return err
	}
	t.uConn=conn
	log.Printf("enabled kernel gtp device %s", gtpInterface)

	go func() {
		log.Printf("about to listen and serve")
		if err := conn.ListenAndServe(ctx); err != nil {
			log.Println(err)
			return
		}
		log.Println("conn.ListenAndServe exited")
	}()
		log.Printf("Started userplane on %s", localAddress)

	if err := conn.AddTunnelOverride(
		gnb.Recv.GTPuPeerAddr, ue.Recv.PDUAddress,
		gnb.Recv.GTPuPeerTEID, gnb.GTPuTEID); err != nil {
		log.Println(err)
		return err
	}
	log.Printf("created tunnel from %s %s",gnb.Recv.GTPuPeerAddr, ue.Recv.PDUAddress)

	time.Sleep(time.Second * 3)

	if err := addRoute(conn); err != nil {
		log.Fatalf("failed to addRoute: %v", err)
		return err
	}

	err := addIP(gnb.GTPuIFname, ue.Recv.PDUAddress, 24)
	if err != nil {
		log.Fatalf("failed to addIP: %v", err)
		return err
	}

	err = addRule(gtpInterface,ue.Recv.PDUAddress)
	if err != nil {
		log.Fatalf("failed to addRule: %v", err)
		return err
	}

	return nil
}

func addRoute(c *gtpv1.UPlaneConn) error{
	log.Printf("addRoute function called")
	log.Printf("routing %s to %s",net.IPv4zero.String(), c.KernelGTP.Link.Attrs().Index)

	route := &netlink.Route{
		Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)}, // default
		LinkIndex: c.KernelGTP.Link.Attrs().Index,                          // dev gtp-<ECI>
		Scope:     netlink.SCOPE_LINK,                                      // scope link
		Protocol:  4,                                                       // proto static
		Priority:  1,                                                       // metric 1
		Table:     1001,                                                    // table <ECI>
	}

	return netlink.RouteReplace(route)
}
	
func addIP(ifname string, ip net.IP, mask int) error{
	log.Printf("addIP function called with %s %v %d",ifname, ip, mask)

	link, err := netlink.LinkByName(ifname)
	if err != nil {
		return err
	}
	addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
	if err != nil {
		return err
	}
	log.Printf("addrs on if %v",addrs)

	netToAdd := &net.IPNet{IP: ip, Mask: net.CIDRMask(24, 32)}
	var addr netlink.Addr
	addr.IPNet = netToAdd
	if err := netlink.AddrAdd(link, &addr); err != nil {
		return err
	}

	return nil
}

func addRule(ifname string, ip net.IP) error{
	log.Printf("addRule function called")

	// rules, err := netlink.RuleList(0)
	// if err != nil {
	// 	return err
	// }

	mask32 := &net.IPNet{IP: ip, Mask: net.CIDRMask(24, 24)}
	rule := netlink.NewRule()
	rule.IifName = ifname
	rule.Src = mask32
	rule.Table = 1001

	return netlink.RuleAdd(rule)
}

func setupUserPlane(t *testSession, ctx context.Context, c *ngap.Camper) error  {
	log.Printf("setupUserPlane function called")

	gnb := t.gnb
	ue := c.UE

	c.GTPu = gtp.NewGTP(gnb.GTPuTEID, gnb.Recv.GTPuPeerTEID)
	gtpu := c.GTPu
	gtpu.SetExtensionHeader(true)
	gtpu.SetQosFlowID(c.QosFlowID)

	log.Printf("GTPuIFname: %s\n", gnb.GTPuIFname)
	log.Printf("GTP-U Peer: %v\n", gnb.Recv.GTPuPeerAddr)
	log.Printf("GTP-U Peer TEID: %v\n", gnb.Recv.GTPuPeerTEID)
	log.Printf("GTP-U Local TEID: %v\n", gnb.GTPuTEID)
	log.Printf("QoS Flow ID: %d\n", gtpu.QosFlowID)
	log.Printf("UE address: %v\n", ue.Recv.PDUAddress)

	if err := runUserPlane(t,gnb.GTPuLocalAddr,gnb.GTPuIFname,ctx,c); err !=nil{
		return err
	}

	return nil
}


func cleanupUserPlane(t *testSession){
	log.Printf("cleanupUserPlane function called")

	if c := t.uConn; c != nil {
		if err := c.Close(); err != nil {
			log.Println(err)
		}
	}
	if c := t.uConn; c != nil {
		if err := c.Close(); err != nil {
			log.Println(err)
		}
	}

	return
}
