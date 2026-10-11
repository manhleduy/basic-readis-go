package main

import (
	//"fmt"
	
	"flag"
	"fmt"
	
	"log"
	"log/slog"
	"net"
	
)


const DefaultListenAddr = ":5001"


type Config struct {
	ListenAddr string
}
type Message struct{
	cmd Command
	peer *Peer
}
type Server struct {
	Config
	peers 		map[*Peer]bool
	ln 			net.Listener
	addPeerch 	chan *Peer
	delPeerch   chan *Peer
	quitCh 		chan struct{}
	msgCh 		chan Message

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
		delPeerch: 	make(chan *Peer),
		quitCh: 	make(chan struct{}),
		msgCh: 		make(chan Message),
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

func (s *Server) handleMessage(msg Message) error {
	
	
	switch v:=msg.cmd.(type){
	case SetCommand:
		return s.kv.Set(v.key, v.val)
	case GetCommand:
		val, ok:= s.kv.Get(v.key)
		if !ok{
			return fmt.Errorf("key not found")
		}
		_, err := msg.peer.Send(val)
		if err != nil{
			slog.Error("peer send error", "err", err)
		}
		
		
		
	}

	return nil
}

func (s *Server) loop(){
	for {
		select{
		case msg := <- s.msgCh:
			if err :=s.handleMessage(msg); err != nil{
				slog.Error("raw message error", "err", err)
				
			}
			
		case <- s.quitCh:
			return
		case peer := <- s.addPeerch:
			slog.Info("new peer connected", "remoteAddr", peer.conn.RemoteAddr())
			s.peers[peer]= true
		case peer := <- s.delPeerch:
			slog.Info("peer disconnected", "remoteAddr", peer.conn.RemoteAddr())

			delete(s.peers, peer)
		
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
	peer := NewPeer(conn, s.msgCh,s.delPeerch);
	s.addPeerch <- peer

	if err := peer.readLoop();err != nil{
		slog.Error("peer read error", "remoteAddr", conn.RemoteAddr())
	}
}
func main(){
	listenAddr := flag.String(
		"listenAddr", 
		DefaultListenAddr, 
		"listen addres of the goredis server",
	)

	flag.Parse()
	
	server := NewServer(Config{
		ListenAddr: *listenAddr,

	})
	log.Fatal(server.Start())
}

