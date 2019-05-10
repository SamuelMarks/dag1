package tester

import (
	"fmt"
	"math/rand"
	_ "os" // required for TODO
	"strconv"
	"strings"
	_ "sync" // required for TODO
	"time"

	"github.com/SamuelMarks/dag1/src/peers"
	"github.com/SamuelMarks/dag1/src/proxy"
	"github.com/sirupsen/logrus"
)

// PingNodesN ping the nodes to make sure they are communicating
func PingNodesN(participants []*peers.Peer, p peers.PubKeyPeers, n uint64, delay uint64, logger *logrus.Logger, ProxyAddr string) {
	// pause before shooting test transactions
	time.Sleep(time.Duration(delay) * time.Second)

	proxies := make(map[uint64]*proxy.GrpcDAG1Proxy)
	for _, participant := range participants {
		node := p[participant.Message.PubKeyHex]
		if node.Message.NetAddr == "" {
			fmt.Printf("node missing NetAddr [%v]", node)
			continue
		}
		hostPort := strings.Split(node.Message.NetAddr, ":")
		port, err := strconv.Atoi(hostPort[1])
		if err != nil {
			fmt.Printf("error:\t\t\t%s\n", err.Error())
			fmt.Printf("Unable to create port:\t\t\t%s (id=%d)\n", participant.Message.NetAddr, node.ID)
		}
		addr := fmt.Sprintf("%s:%d", hostPort[0], port-3000 /*9000*/)
		dag1Proxy, err := proxy.NewGrpcDAG1Proxy(addr, logger)
		if err != nil {
			fmt.Printf("error:\t\t\t%s\n", err.Error())
			fmt.Printf("Failed to create WebsocketDAG1Proxy:\t\t\t%s (id=%d)\n", participant.Message.NetAddr, node.ID)
		}
		proxies[node.ID] = dag1Proxy
	}
	for iteration := uint64(0); iteration < n; iteration++ {
		participant := participants[rand.Intn(len(participants))]
		node := p[participant.Message.PubKeyHex]

		_, err := transact(proxies[node.ID], ProxyAddr, iteration)

		if err != nil {
			fmt.Printf("error:\t\t\t%s\n", err.Error())
			fmt.Printf("Failed to ping:\t\t\t%s (id=%d)\n", participant.Message.NetAddr, node.ID)
			fmt.Printf("Failed to send transaction:\t%d\n\n", iteration)
		} /*else {
			fmt.Printf("Pinged:\t\t\t%s (id=%d)\n", participant.NetAddr, node)
			fmt.Printf("Last transaction sent:\t%d\n\n", iteration)
		}*/
	}

	for _, dag1Proxy := range proxies {
		if err := dag1Proxy.Close(); err != nil {
			logger.Fatal(err)
		}
	}
	fmt.Println("Pinging stopped after ", n, " iterations")
}

func transact(proxy *proxy.GrpcDAG1Proxy, proxyAddr string, iteration uint64) (string, error) {

	// Ethereum txns are ~108 bytes. Bitcoin txns are ~250 bytes.
	// A good assumption is to make txns 120 bytes in size.
	// However, for speed, we're using 1 byte here. Modify accordingly.
	msg := []byte{ 0 }
	for i := 0; i < 10; i++ {
		// Send 10 txns to the server.
		//msg := fmt.Sprintf("%s.%d.%d", proxyAddr, iteration, i)
		//err := proxy.SubmitTx([]byte(msg))
		err := proxy.SubmitTx(msg)
		if err != nil {
			return "", err
		}
	}
	// fmt.Println("Submitted tx, ack=", ack)  # `ack` is now `_`

	return "", nil
}
