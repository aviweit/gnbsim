package main

import (
	"log"
	"os"
	"os/signal"
	"context"
	"syscall"
)

func main() {
	log.SetPrefix("[5g-gateway]")
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	t := initRAN()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	// signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGCONT)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	log.Printf("run user plane")
	t.updateNGAP("gnb.json")
	t.initUE()

    // register all UEs
	gnb := t.gnb
	for _, c := range gnb.Camper {
		ue := c.UE
		t.registerUE(ue)
	}

	for _, c := range gnb.Camper {
		ue := c.UE
		t.establishPDUSession(ue)
	}
	log.Printf("before user plane setup")

	fatalCh := make(chan error, 1)

	for _, c := range gnb.Camper {
		go func() {
			if err := setupUserPlane(t, ctx,c); err != nil {
				fatalCh <- err
			}
		}()
	}

	for {
		select {
		case err := <-fatalCh:
			log.Printf("FATAL: %s", err)
			return
		}
	}

	return
}

