package labrpc

//
// channel-based RPC, for github.com/ArshiAbolghasemi/disgo labs.
//
// simulates a network that can lose requests, lose replies,
// delay messages, and entirely disconnect particular hosts.
//
// we will use the original labrpc.go to test your code for grading.
// so, while you can modify this code to help you debug, please
// test against the original before submitting.
//
// adapted from Go net/rpc/server.go.
//
// sends labgob-encoded values to ensure that RPCs
// don't include references to program objects.
//
// net := MakeNetwork() -- holds network, clients, servers.
// end := net.MakeEnd(endname) -- create a client end-point, to talk to one server.
// net.AddServer(servername, server) -- adds a named server to network.
// net.DeleteServer(servername) -- eliminate the named server.
// net.Connect(endname, servername) -- connect a client to a server.
// net.Enable(endname, enabled) -- enable/disable a client.
// net.Reliable(bool) -- false means drop/delay messages
//
// end.Call("Raft.AppendEntries", &args, &reply) -- send an RPC, wait for reply.
// the "Raft" is the name of the server struct to be called.
// the "AppendEntries" is the name of the method to be called.
// Call() returns true to indicate that the server executed the request
// and the reply is valid.
// Call() returns false if the network lost the request or reply
// or the server is down.
// It is OK to have multiple Call()s in progress at the same time on the
// same ClientEnd.
// Concurrent calls to Call() may be delivered to the server out of order,
// since the network may re-order messages.
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.
// the server RPC handler function must declare its args and reply arguments
// as pointers, so that their types exactly match the types of the arguments
// to Call().
//
// srv := MakeServer()
// srv.AddService(svc) -- a server can have multiple services, e.g. Raft and k/v
//   pass srv to net.AddServer()
//
// svc := MakeService(receiverObject) -- obj's methods will handle RPCs
//   much like Go's rpcs.Register()
//   pass svc to srv.AddService()
//

import (
	"bytes"
	"log"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ArshiAbolghasemi/disgo/labgob"
)

const (
	SHORTDELAY = 27   // ms
	LONGDELAY  = 7000 // ms
	MAXDELAY   = LONGDELAY + 100
)

type reqMsg struct {
	endname  interface{} // name of sending ClientEnd
	svcMeth  string      // e.g. "Raft.AppendEntries"
	argsType reflect.Type
	args     []byte
	replyCh  chan replyMsg
}

type replyMsg struct {
	ok    bool
	reply []byte
}

type ClientEnd struct {
	endname interface{}   // this end-point's name
	ch      chan reqMsg   // copy of Network.endCh
	done    chan struct{} // closed when Network is cleaned up
}

// send an RPC, wait for the reply.
// the return value indicates success; false means that
// no reply was received from the server.
func (e *ClientEnd) Call(svcMeth string, args interface{}, reply interface{}) bool {
	req := reqMsg{}
	req.endname = e.endname
	req.svcMeth = svcMeth
	req.argsType = reflect.TypeOf(args)
	req.replyCh = make(chan replyMsg)

	qb := new(bytes.Buffer)
	qe := labgob.NewEncoder(qb)
	if err := qe.Encode(args); err != nil {
		panic(err)
	}
	req.args = qb.Bytes()

	//
	// send the request.
	//
	select {
	case e.ch <- req:
		// the request has been sent.
	case <-e.done:
		// entire Network has been destroyed.
		return false
	}

	//
	// wait for the reply.
	//
	rep := <-req.replyCh
	if rep.ok {
		rb := bytes.NewBuffer(rep.reply)
		rd := labgob.NewDecoder(rb)
		if err := rd.Decode(reply); err != nil {
			log.Fatalf("ClientEnd.Call(): decode reply: %v\n", err)
		}
		return true
	} else {
		return false
	}
}

type Network struct {
	mu             sync.Mutex
	reliable       bool
	longDelays     bool                        // pause a long time on send on disabled connection
	longReordering bool                        // sometimes delay replies a long time
	ends           map[interface{}]*ClientEnd  // ends, by name
	enabled        map[interface{}]bool        // by end name
	servers        map[interface{}]*Server     // servers, by name
	connections    map[interface{}]interface{} // endname -> servername
	endCh          chan reqMsg
	done           chan struct{} // closed when Network is cleaned up
	count          int32         // total RPC count, for statistics
	bytes          int64         // total bytes send, for statistics
}

func MakeNetwork() *Network {
	rn := &Network{}
	rn.reliable = true
	rn.ends = map[interface{}]*ClientEnd{}
	rn.enabled = map[interface{}]bool{}
	rn.servers = map[interface{}]*Server{}
	rn.connections = map[interface{}](interface{}){}
	rn.endCh = make(chan reqMsg)
	rn.done = make(chan struct{})

	// single goroutine to handle all ClientEnd.Call()s
	go func() {
		for {
			select {
			case xreq := <-rn.endCh:
				atomic.AddInt32(&rn.count, 1)
				atomic.AddInt64(&rn.bytes, int64(len(xreq.args)))
				go rn.processReq(xreq)
			case <-rn.done:
				return
			}
		}
	}()

	return rn
}

func (rn *Network) Cleanup() {
	close(rn.done)
}

func (rn *Network) Reliable(yes bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.reliable = yes
}

func (rn *Network) IsReliable() bool {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	return rn.reliable
}

func (rn *Network) LongReordering(yes bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.longReordering = yes
}

func (rn *Network) LongDelays(yes bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.longDelays = yes
}

func (rn *Network) IsLongDelays() bool {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	return rn.longDelays
}

func (rn *Network) readEndnameInfo(endname interface{}) (enabled bool,
	servername interface{}, server *Server, reliable bool, longreordering bool,
) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	enabled = rn.enabled[endname]
	servername = rn.connections[endname]
	if servername != nil {
		server = rn.servers[servername]
	}
	reliable = rn.reliable
	longreordering = rn.longReordering
	return
}

func (rn *Network) isServerDead(endname interface{}, servername interface{}, server *Server) bool {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if rn.enabled[endname] == false || rn.servers[servername] != server {
		return true
	}
	return false
}

func (rn *Network) processReq(req reqMsg) {
	enabled, servername, server, reliable, longreordering := rn.readEndnameInfo(req.endname)

	if enabled && servername != nil && server != nil {
		if reliable == false {
			// short delay
			ms := (rand.Int() % SHORTDELAY)
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}

		if reliable == false && (rand.Int()%1000) < 100 {
			// drop the request, return as if timeout
			req.replyCh <- replyMsg{false, nil}
			return
		}

		// execute the request (call the RPC handler).
		// in a separate thread so that we can periodically check
		// if the server has been killed and the RPC should get a
		// failure reply.
		ech := make(chan replyMsg)
		go func() {
			r := server.dispatch(req)
			ech <- r
		}()

		// wait for handler to return,
		// but stop waiting if DeleteServer() has been called,
		// and return an error.
		var reply replyMsg
		replyOK := false
		serverDead := false
		for replyOK == false && serverDead == false {
			select {
			case reply = <-ech:
				replyOK = true
			case <-time.After(100 * time.Millisecond):
				serverDead = rn.isServerDead(req.endname, servername, server)
				if serverDead {
					go func() {
						<-ech // drain channel to let the goroutine created earlier terminate
					}()
				}
			}
		}

		// do not reply if DeleteServer() has been called, i.e.
		// the server has been killed. this is needed to avoid
		// situation in which a client gets a positive reply
		// to an Append, but the server persisted the update
		// into the old Persister. config.go is careful to call
		// DeleteServer() before superseding the Persister.
		serverDead = rn.isServerDead(req.endname, servername, server)

		if replyOK == false || serverDead == true {
			// server was killed while we were waiting; return error.
			req.replyCh <- replyMsg{false, nil}
		} else if reliable == false && (rand.Int()%1000) < 100 {
			// drop the reply, return as if timeout
			req.replyCh <- replyMsg{false, nil}
		} else if longreordering == true && rand.Intn(900) < 600 {
			// delay the response for a while
			ms := 200 + rand.Intn(1+rand.Intn(2000))
			// Russ points out that this timer arrangement will decrease
			// the number of goroutines, so that the race
			// detector is less likely to get upset.
			time.AfterFunc(time.Duration(ms)*time.Millisecond, func() {
				atomic.AddInt64(&rn.bytes, int64(len(reply.reply)))
				req.replyCh <- reply
			})
		} else {
			atomic.AddInt64(&rn.bytes, int64(len(reply.reply)))
			req.replyCh <- reply
		}
	} else {
		// simulate no reply and eventual timeout.
		ms := 0
		if rn.IsLongDelays() {
			// let Raft tests check that leader doesn't send
			// RPCs synchronously.
			ms = (rand.Int() % LONGDELAY)
		} else {
			// many kv tests require the client to try each
			// server in fairly rapid succession.
			ms = (rand.Int() % 100)
		}
		time.AfterFunc(time.Duration(ms)*time.Millisecond, func() {
			req.replyCh <- replyMsg{false, nil}
		})
	}

}

// create a client end-point.
// start the thread that listens and delivers.
func (rn *Network) MakeEnd(endname interface{}) *ClientEnd {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if _, ok := rn.ends[endname]; ok {
		log.Fatalf("MakeEnd: %v already exists\n", endname)
	}

	e := &ClientEnd{}
	e.endname = endname
	e.ch = rn.endCh
	e.done = rn.done
	rn.ends[endname] = e
	rn.enabled[endname] = false
	rn.connections[endname] = nil

	return e
}

func (rn *Network) DeleteEnd(endname interface{}) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if _, ok := rn.ends[endname]; !ok {
		log.Fatalf("MakeEnd: %v doesn't exists\n", endname)
	}
	delete(rn.ends, endname)
	delete(rn.enabled, endname)
	delete(rn.connections, endname)
}

func (rn *Network) AddServer(servername interface{}, rs *Server) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.servers[servername] = rs
}

func (rn *Network) DeleteServer(servername interface{}) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.servers[servername] = nil
}

// connect a ClientEnd to a server.
// a ClientEnd can only be connected once in its lifetime.
func (rn *Network) Connect(endname interface{}, servername interface{}) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.connections[endname] = servername
}

// enable/disable a ClientEnd.
func (rn *Network) Enable(endname interface{}, enabled bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.enabled[endname] = enabled
}

// get a server's count of incoming RPCs.
func (rn *Network) GetCount(servername interface{}) int {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	svr := rn.servers[servername]
	return svr.GetCount()
}

func (rn *Network) GetTotalCount() int {
	x := atomic.LoadInt32(&rn.count)
	return int(x)
}

func (rn *Network) GetTotalBytes() int64 {
	x := atomic.LoadInt64(&rn.bytes)
	return x
}

// a server is a collection of services, all sharing
// the same rpc dispatcher. so that e.g. both a Raft
// and a k/v server can listen to the same rpc endpoint.
type Server struct {
	mu       sync.Mutex
	services map[string]*Service
	count    int // incoming RPCs
}

func MakeServer() *Server {
	rs := &Server{}
	rs.services = map[string]*Service{}
	return rs
}

func (rs *Server) AddService(svc *Service) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.services[svc.name] = svc
}

func (rs *Server) dispatch(req reqMsg) replyMsg {
	rs.mu.Lock()

	rs.count += 1

	// split Raft.AppendEntries into service and method
	dot := strings.LastIndex(req.svcMeth, ".")
	serviceName := req.svcMeth[:dot]
	methodName := req.svcMeth[dot+1:]

	service, ok := rs.services[serviceName]

	rs.mu.Unlock()

	if ok {
		return service.dispatch(methodName, req)
	} else {
		choices := []string{}
		for k, _ := range rs.services {
			choices = append(choices, k)
		}
		log.Fatalf("labrpc.Server.dispatch(): unknown service %v in %v.%v; expecting one of %v\n",
			serviceName, serviceName, methodName, choices)
		return replyMsg{false, nil}
	}
}

func (rs *Server) GetCount() int {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.count
}

// an object with methods that can be called via RPC.
// a single server may have more than one Service.
type Service struct {
	name    string
	rcvr    reflect.Value
	typ     reflect.Type
	methods map[string]reflect.Method
}

func MakeService(rcvr interface{}) *Service {
	svc := &Service{}
	svc.typ = reflect.TypeOf(rcvr)
	svc.rcvr = reflect.ValueOf(rcvr)
	svc.name = reflect.Indirect(svc.rcvr).Type().Name()
	svc.methods = map[string]reflect.Method{}

	for m := 0; m < svc.typ.NumMethod(); m++ {
		method := svc.typ.Method(m)
		mtype := method.Type
		mname := method.Name

		//fmt.Printf("%v pp %v ni %v 1k %v 2k %v no %v\n",
		//	mname, method.PkgPath, mtype.NumIn(), mtype.In(1).Kind(), mtype.In(2).Kind(), mtype.NumOut())

		if method.PkgPath != "" || // capitalized?
			mtype.NumIn() != 3 ||
			//mtype.In(1).Kind() != reflect.Ptr ||
			mtype.In(2).Kind() != reflect.Ptr ||
			mtype.NumOut() != 0 {
			// the method is not suitable for a handler
			//fmt.Printf("bad method: %v\n", mname)
		} else {
			// the method looks like a handler
			svc.methods[mname] = method
		}
	}

	return svc
}

func (svc *Service) dispatch(methname string, req reqMsg) replyMsg {
	if method, ok := svc.methods[methname]; ok {
		// prepare space into which to read the argument.
		// the Value's type will be a pointer to req.argsType.
		args := reflect.New(req.argsType)

		// decode the argument.
		ab := bytes.NewBuffer(req.args)
		ad := labgob.NewDecoder(ab)
		ad.Decode(args.Interface())

		// allocate space for the reply.
		replyType := method.Type.In(2)
		replyType = replyType.Elem()
		replyv := reflect.New(replyType)

		// call the method.
		function := method.Func
		function.Call([]reflect.Value{svc.rcvr, args.Elem(), replyv})

		// encode the reply.
		rb := new(bytes.Buffer)
		re := labgob.NewEncoder(rb)
		re.EncodeValue(replyv)

		return replyMsg{true, rb.Bytes()}
	} else {
		choices := []string{}
		for k, _ := range svc.methods {
			choices = append(choices, k)
		}
		log.Fatalf("labrpc.Service.dispatch(): unknown method %v in %v; expecting one of %v\n",
			methname, req.svcMeth, choices)
		return replyMsg{false, nil}
	}
}
