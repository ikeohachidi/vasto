package store

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/chrislusf/vasto/pb"
	"github.com/chrislusf/vasto/util"
	"github.com/golang/protobuf/proto"
)

// Run starts the heartbeating to master and starts accepting requests.
func (ss *storeServer) serveTcp(listener net.Listener) {

	for {
		// Listen for an incoming connection.
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			continue
		}
		// Handle connections in a new goroutine.
		go func() {
			if err = conn.SetDeadline(time.Time{}); err != nil {
				fmt.Printf("Failed to set timeout: %v\n", err)
			}
			if c, ok := conn.(*net.TCPConn); ok {
				c.SetKeepAlive(true)
				c.SetNoDelay(true)
				c.SetLinger(0)
			}
			ss.handleConnection(conn)
			conn.Close()
		}()
	}
}

func (ss *storeServer) handleConnection(conn net.Conn) {

	reader := bufio.NewReader(conn)

	for {
		if err := ss.handleRequest(reader, conn); err != nil {
			if err != io.EOF {
				log.Printf("handleRequest: %v", err)
			}
			return
		}
	}

}

func (ss *storeServer) handleRequest(reader io.Reader, writer io.Writer) error {

	var input, output []byte
	var err error

	input, err = util.ReadMessage(reader)

	if err == io.EOF {
		return err
	}
	if err != nil {
		return fmt.Errorf("read message: %v", err)
	}

	requests := &pb.Requests{}
	if err = proto.Unmarshal(input, requests); err != nil {
		return fmt.Errorf("unmarshal: %v", err)
	}

	responses := &pb.Responses{}
	for _, request := range requests.Requests {
		response := ss.processRequest(request)
		responses.Responses = append(responses.Responses, response)
	}

	output, err = proto.Marshal(responses)
	if err != nil {
		return fmt.Errorf("marshal: %v", err)
	}

	err = util.WriteMessage(writer, output)
	if err != nil {
		return fmt.Errorf("write message: %v", err)
	}

	return nil

}

func (ss *storeServer) processRequest(command *pb.Request) *pb.Response {
	if command.GetGet() != nil {
		return &pb.Response{
			Get: ss.processGet(command.Get),
		}
	} else if command.GetPut() != nil {
		return &pb.Response{
			Put: ss.processPut(command.Put),
		}
	} else if command.GetDelete() != nil {
		return &pb.Response{
			Delete: ss.processDelete(command.Delete),
		}
	} else if command.GetGetByPrefix() != nil {
		return &pb.Response{
			GetByPrefix: ss.processPrefix(command.GetByPrefix),
		}
	}
	return &pb.Response{
		Put: &pb.PutResponse{
			Ok: true,
		},
	}
}
