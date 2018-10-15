package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
	//	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
)

const BUFF_SIZE = 1024
const MAX_SIZE_EXPECT = 1024

type proxySt struct {
	connMap      map[int]*net.TCPConn
	sessid       int
	masterSessid int
	msg_chan     chan string
	sync.RWMutex
}

var g_mst *proxySt

func (this *proxySt) incSess() {
	this.Lock()
	defer this.Unlock()
	this.sessid = this.sessid + 1
}

func (this *proxySt) setMasterId(id int) {
	this.Lock()
	defer this.Unlock()
	this.masterSessid = id
}

func (this *proxySt) RemConn(sessid int) {
	this.Lock()
	defer this.Unlock()
	delete(this.connMap, sessid)
	if sessid == this.masterSessid {
		this.masterSessid = 0
	}
}
func (this *proxySt) AddConn(conn *net.TCPConn) int {
	this.Lock()
	defer this.Unlock()
	this.sessid++
	this.connMap[this.sessid] = conn
	return this.sessid
}

func (this *proxySt) getMasterConn() *net.TCPConn {
	this.Lock()
	defer this.Unlock()
	ret, ok := this.connMap[this.masterSessid]
	if !ok {
		return nil
	}
	return ret
}

func handleConn(tcpConn *net.TCPConn, sessid int) {
	if tcpConn == nil {
		return
	}

	head := make([]byte, 4)
	hsize := uint32(0)
	for {
		if _, err := io.ReadFull(tcpConn, head); err != nil {
			log.Println(err)
			tcpConn.Close()
			break
		}
		hbuf := bytes.NewReader(head)
		if err := binary.Read(hbuf, binary.LittleEndian, &hsize); err != nil {
			log.Println(err)
			tcpConn.Close()
			break
		}
		if hsize > MAX_SIZE_EXPECT {
			log.Println("nowsize:", hsize, ", exceed max:", MAX_SIZE_EXPECT)
			tcpConn.Close()
			break
		}
		membuf := make([]byte, hsize)
		if _, err := io.ReadFull(tcpConn, membuf); err != nil {
			log.Println(err)
			tcpConn.Close()
			break
		}

		recvStr := string(membuf)
		log.Println("recv content:", recvStr, ", ssid:", sessid)
		if recvStr == "master" {
			g_mst.setMasterId(sessid)
		} else {
			log.Println("get conn sessid: ", sessid, " oper:", recvStr)
		}
	}
	g_mst.RemConn(sessid)
	log.Println("conn closed, sessid: ", sessid)
}

// 错误处理
func handleError(err error) {
	if err == nil {
		return
	}
	log.Println("error:%s\n", err.Error())
}

func handleMsgRoute() {
	go func() {
		for {
			select {
			case val := <-g_mst.msg_chan:
				conn := g_mst.getMasterConn()
				if conn == nil {
					time.Sleep(time.Second)
					g_mst.msg_chan <- val
					continue
				}
				ssize := len(val)
				bout := bytes.NewBuffer([]byte{})
				binary.Write(bout, binary.LittleEndian, uint32(ssize))
				binary.Write(bout, binary.LittleEndian, []byte(val))
				n, e := conn.Write(bout.Bytes())
				log.Println("handleMsgRoute n:", n, ",e:", e)
			}
		}
	}()
}

func handleHttp(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(BUFF_SIZE)
	if len(r.Form) == 0 {
		return
	}
	jval, _ := json.Marshal(r.Form)
	jval_str := string(jval)
	log.Println("handleHttp data:", jval_str)
	g_mst.msg_chan <- jval_str

	fmt.Fprintln(w, string(jval))
	r.Body.Close()
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage:<command> <sport> <webport>")
		return
	}
	port := os.Args[1]
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":"+port)
	if err != nil {
		handleError(err)
	}
	tcpListener, err := net.ListenTCP("tcp4", tcpAddr) //监听
	if err != nil {
		handleError(err)
	}
	defer tcpListener.Close()

	log.Println("listen tcp on :", port)
	g_mst = &proxySt{
		connMap:      make(map[int]*net.TCPConn),
		sessid:       0,
		masterSessid: 0,
		msg_chan:     make(chan string, 1024),
	}
	go func() {
		for {
			tcpConn, err := tcpListener.AcceptTCP()
			sessid := g_mst.AddConn(tcpConn)
			log.Println(fmt.Sprintf("The client:%s has connected!\n", tcpConn.RemoteAddr().String()))
			if err != nil {
				handleError(err)
			}
			defer tcpConn.Close()
			go handleConn(tcpConn, sessid) //起一个goroutine处理
		}
	}()

	go handleMsgRoute()
	log.Println("listen http on :", os.Args[2])
	//start http server
	http.HandleFunc("/", handleHttp)
	webPort := fmt.Sprintf(":%s", os.Args[2])
	http.ListenAndServe(webPort, nil)

}
