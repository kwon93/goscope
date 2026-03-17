package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"github.com/kwon93/goscope/internal/capture"
	"github.com/kwon93/goscope/internal/output"
)

func main() {
	// 1. flag 파싱 (iface, filter 정도면 충분)
	ifaces := flag.String("i", "", "네트워크 인터페이스")
	filter := flag.String("f", "", "BPF 필터 (ex: tcp port 80)")
	outFile := flag.String("w", "", "저장할 pcap 파일명 (ex: capture.pcap)")
	var writer *pcapgo.Writer

	flag.Parse()
	if *outFile == "" {
		fmt.Print("저장할 파일명을 입력하세요 (생략시 Enter): ")
		name, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		name = strings.TrimSpace(name) // 줄바꿈 제거
		if name != "" && !strings.HasSuffix(name, ".pcap") {
			name = name + ".pcap"
		}
		*outFile = name
	}
	if *outFile != "" {
		f, err := os.Create(*outFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "파일 생성 실패: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		writer = pcapgo.NewWriter(f)
		writer.WriteFileHeader(65536, layers.LinkTypeEthernet)
	}

	if *ifaces == "" {
		devices, err := pcap.FindAllDevs()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		for i, d := range devices {
			fmt.Printf("[%d] %s (%s)\n", i, d.Description, d.Name)
		}

		var num int
		fmt.Scan(&num)
		*ifaces = devices[num].Name
	}

	// 2. Engine 생성 (WithInterface, WithFilter...)
	engine := capture.NewEngine(
		capture.WithInterface(*ifaces),
		capture.WithFilter(*filter),
	)
	// 3. ctx, cancel (signal.NotifyContext)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// 4. Start()
	packets, err := engine.Start(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 5. for range packets → output.Print()
	fmt.Println("패킷 캡쳐 시작 (Ctrl+C를 눌러 종료하세요)")
	for packet := range packets {
		output.Print(packet)

		if writer != nil {
			writer.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
		}
	}
}
