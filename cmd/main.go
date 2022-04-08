package main

//Исходники задания для первого занятия у других групп https://github.com/t0pep0/GB_best_go

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"practic/discovery"
	"practic/reader"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	discovery := discovery.NewDiscovery(ctx, &reader.FSReader{})
	const wantExt = ".go"

	defer cancel()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	waitCh := make(chan struct{})
	go func() {

		res, err := discovery.FindFiles(wantExt)
		if err != nil {
			log.Printf("Error on search: %v\n", err)
			os.Exit(1)
		}
		for _, f := range res {
			fmt.Printf("\tName: %s\t\t Path: %s\n", f.Name, f.Path)
		}
		waitCh <- struct{}{}
	}()
	go func() {
		<-sigCh
		log.Println("Signal received, terminate...")
		cancel()
	}()

	<-waitCh
	log.Println("Done")
}
