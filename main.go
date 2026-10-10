package main

import (
	//"fmt"
	"context"
	"fmt"
	"goredis/client"
	"log"
	"log/slog"
	"net"
	"time"
)


const DefaultListenAddr = ":5001"


type Config struct {
	ListenAddr string
}
type Server struct {
	Config
	peers 		map[*Peer]bool
	ln 			net.Listener
	addPeerch 	chan *Peer
	quitCh 		chan struct{}
	msgCh 		chan []byte
	kv			*KV

}

func NewServer(cfg Config) *Server {
	if len(cfg.ListenAddr)==0{
		cfg.ListenAddr= DefaultListenAddr
	}
	return &Server{
		Config: 	cfg,
		peers: 		make(map[*Peer]bool),
		addPeerch: 	make(chan *Peer),   
		quitCh: 	make(chan struct{}),
		msgCh: 		make(chan []byte),
		kv:			NewKV(),
	}
}
func (s *Server) Start() error{
	ln , err:= net.Listen("tcp", s.ListenAddr)
	if err != nil{
		return err
	}
	s.ln = ln

	go s.loop()
	slog.Info("server running", "listenAddr", s.ListenAddr)
	s.acceptLoop()
	return nil
}

func (s *Server) handleRawMessage(rawMsg []byte) error {
	cmd, err := parseCommand(string(rawMsg))
	if err != nil{
		return err
	}
	
	switch v:=cmd.(type){
	case SetCommand:
		return s.kv.Set(v.key, v.val)
	}
	return nil
}

func (s *Server) loop(){
	for {
		select{
		case rawMsg := <- s.msgCh:
			if err :=s.handleRawMessage(rawMsg); err != nil{
				slog.Error("raw message error", "err", err)
				
			}
			fmt.Println(rawMsg)
		case <- s.quitCh:
			return
		case peer := <- s.addPeerch:
			s.peers[peer]= true
		
	}
}
}
func (s *Server) acceptLoop() error {
	for{
		conn, err := s.ln.Accept()
		if err != nil{
			slog.Error("accept error", "err", err)
			continue
		}
		go s.handleConn(conn)
	}


}
func (s *Server) handleConn(conn net.Conn){
	peer := NewPeer(conn, s.msgCh);
	s.addPeerch <- peer
	slog.Info("new peer connected", "remoeAddr", conn.RemoteAddr())
	if err := peer.readLoop();err != nil{
		slog.Error("peer read error", "remoteAddr", conn.RemoteAddr())
	}
}
func main(){
	server := NewServer(Config{})

	go func(){
		log.Fatal(server.Start())
	}()
	time.Sleep(time.Second)
	for i :=0; i< 10; i++{
		c := client.New("localhost:5001")
		
		if err := c.Set(context.TODO(),fmt.Sprintf("bar %d", i), fmt.Sprintf("bar %d", i)); err != nil{
			log.Fatal(err)
		}

		
	//select {} // we are blocking here so the program doesnot exist
	}
	fmt.Println(server.kv.data)
	time.Sleep(time.Second)
}