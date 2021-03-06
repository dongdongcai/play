package raftkv

import "6.824/src/labrpc"
import "crypto/rand"
import "math/big"

type Clerk struct {
	servers  []*labrpc.ClientEnd
	id       int64
	opnumber int
	leader   int
	// You will have to modify this struct.
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.servers = servers
	ck.id = nrand()
	ck.opnumber = 0
	ck.leader = 0
	// You'll have to add code here.
	return ck
}

//
// fetch the current value for a key.
// returns "" if the key does not exist.
// keeps trying forever in the face of all other errors.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("RaftKV.Get", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
//
func (ck *Clerk) Get(key string) string {
	ck.opnumber += 1
	request := GetArgs{Key: key, ClientID: ck.id, OpNum: ck.opnumber}
	var reply GetReply
	reply.WrongLeader = true
	for reply.Err == "" {
		for i := 0; i < len(ck.servers); {
			DPrintf("client:Getting %s via server %d", key, (ck.leader+i)%len(ck.servers))
			ok := ck.servers[(ck.leader+i)%len(ck.servers)].Call("RaftKV.Get", &request, &reply)
			if ok && reply.Err != "" {
				ck.leader = (ck.leader + i) % len(ck.servers)
				break
			}
			if reply.WrongLeader {
				i++
			}
		}
	}
	if reply.Err == ErrNoKey {
		return ""
	}
	return reply.Value
}

//
// shared by Put and Append.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("RaftKV.PutAppend", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
//
func (ck *Clerk) PutAppend(key string, value string, op string) {
	ck.opnumber += 1
	request := PutAppendArgs{Key: key, Value: value, Op: op, ClientID: ck.id, OpNum: ck.opnumber}
	var reply PutAppendReply
	reply.WrongLeader = true
	for reply.Err == "" {
		for i := 0; i < len(ck.servers); {
			DPrintf("client:%s %s to %s via server %d", op, value, key, (ck.leader+i)%len(ck.servers))
			ok := ck.servers[(ck.leader+i)%len(ck.servers)].Call("RaftKV.PutAppend", &request, &reply)
			if ok && reply.Err == OK {
				ck.leader = (ck.leader + i) % len(ck.servers)
				break
			}
			if reply.WrongLeader {
				i++
			}
		}
	}
}

func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, "Put")
}
func (ck *Clerk) Append(key string, value string) {
	ck.PutAppend(key, value, "Append")
}
