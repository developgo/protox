/* MIT License
*
* Copyright (c) 2018 Mike Taghavi <mitghi[at]gmail.com>
*
* Permission is hereby granted, free of charge, to any person obtaining a copy
* of this software and associated documentation files (the "Software"), to deal
* in the Software without restriction, including without limitation the rights
* to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
* copies of the Software, and to permit persons to whom the Software is
* furnished to do so, subject to the following conditions:
* The above copyright notice and this permission notice shall be included in all
* copies or substantial portions of the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
* IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
* FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
* LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
* OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
* SOFTWARE.
 */

package networking

import (
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/mitghi/protox/protobase"
	"github.com/mitghi/protox/protocol"
)

type COnline struct {
	constate

	Conn *CLBConnection
}

// MARK: COnline

// NewOnline returns a pointer to a new `Online` struct. This is the where
// interactions with a connected/authorized client happens.
func NewCOnline(conn *CLBConnection) *COnline {
	var (
		co *COnline = &COnline{
			constate: constate{
				constatebase: constatebase{
					Conn: conn,
				},
				client: nil,
				server: nil,
			},
			Conn: conn,
		}
	)
	co.client = conn.GetClient()
	return co
}

// HandleDefault is the default handler ( stub for COnline ).
func (co *COnline) HandleDefault(packet protobase.PacketInterface) (status bool) {
	// NOP
	return true
}

// Section: proto event handler [ event]

// onCONNECT is not valid in this stage.
func (co *COnline) OnCONNECT(packet protobase.PacketInterface) {
	// violates client/server policy
	// terminate
	co.Shutdown()
	co.Conn.protocon.Conn.Close()
}

// onCONNACK is not valid in this stage.
func (co *COnline) OnCONNACK(packet protobase.PacketInterface) {
	// TODO
	const fn string = "onCONNACK"
	logger.FWarn(fn, "-* [COnline] this routine is unimplemented.")
}

// onPUBLISH is the handler for `Publish` packets.
func (co *COnline) OnPUBLISH(packet protobase.PacketInterface) {
	const fn string = "OnPUBLISH"
	var (
		publish *Publish = NewPublish(packet)
		puback  *Puback
		pckt    *Packet // final packet
		pb      *protocol.MsgBox
		pbc     protobase.MsgInterface // publish box clone
	)
	if publish == nil {
		logger.Debug(fn, "- [DecodeErr(COnline)] Unable to decode data into publish packet. error:", publish)
		co.Shutdown()
		return
	}
	// if err := publish.DecodeFrom(packet.GetData()); err != nil {
	// 	logger.Debug(fn, "- [DecodeErr(COnline)] Unable to decode data into publish packet. error:", err)
	// 	co.Shutdown()
	// 	return
	// }
	if stat := co.Conn.storage.AddInbound(publish); stat == false {
		logger.Debug(fn, "- [COnline][NOTICE] addinbound returned false (conline/publish).")
		co.Shutdown()
		return
	}
	puback = NewRawPuback()
	if publish.Meta.Qos > 0 {
		puback.Meta.Qos, puback.Meta.MessageId = publish.Meta.Qos, publish.Meta.MessageId
		if err := puback.Encode(); err != nil {
			logger.FError(fn, "- [COnline] error while encoding puback. error:", err)
			co.Shutdown()
			return
		}
		logger.FTracef(1, fn, "* [QoS] packet QoS(%b) Duplicate(%t) MessageID(%d).", publish.Meta.Qos, publish.Meta.Dup, int(publish.Meta.MessageId))
		pckt = puback.GetPacket().(*Packet)
		logger.FTrace(1, fn, "* [COnline][PubAck] sending packet with content", pckt.Data)
		co.Conn.SendPrio(pckt)
		if stat := co.Conn.storage.DeleteIn(publish); stat == false {
			logger.Debug(fn, "- [COnline][NOTICE] deleteinbound returned false (conline/publish).")
		}
	}
	pb = protocol.NewMsgBox(publish.Meta.Qos, publish.Meta.MessageId,
		protobase.MDInbound, protocol.NewMsgEnvelope(publish.Topic, publish.Message))
	pbc = pb.Clone(protobase.MDInbound)
	/* d e b u g */
	// NOTE
	// . this has changed
	// if publish.Meta.Qos > 0 {
	// 	co.Conn.clblock.Lock()
	// 	// because publish is received when a subscribtion
	// 	// for rotue exists.
	// 	callback, ok := co.Conn.clbsub[puback.Meta.MessageId]
	// 	if ok {
	// 		delete(co.Conn.clbsub, puback.Meta.MessageId)
	// 	}
	// 	co.Conn.clblock.Unlock()
	// 	if ok && callback != nil {
	// 		go func() {
	// 			callback(nil, pbc)
	// 			co.client.Publish(pbc)
	// 		}()
	// 		return
	// 	}
	// }
	// if publish.Meta.Qos > 0 {
	// 	if ok && callback != nil {
	// 		go func() {
	// 			co.client.Publish(pbc)
	// 		}()
	// 		return
	// 	}
	// }
	// TODO
	// . send this to a worker thread
	// go func() { co.client.Publish(pbc) }()
	/* d e b u g */
	co.client.Publish(pbc)
}

// onSUBSCRIBE is the handler for `Subscribe` packets.
func (co *COnline) OnSUBSCRIBE(packet protobase.PacketInterface) {
	logger.Debug("* [Subscribe] packet is received.")
	/* d e b u g */
	// subscribe := NewSubscribe()
	// if err := subscribe.DecodeFrom(packet.Data); err != nil {
	// 	co.Shutdown()
	// 	return
	// }
	// pb := NewMsgBox(subscribe.Meta.Qos, protobase.MDInbound, NewMsgEnvelope(subscribe.Topic, nil))
	// co.client.Subscribe(pb)
	// co.server.NotifySubscribe(co.Conn, pb)
	// co.client.Subscribe(subscribe.Topic)
	// co.server.NotifySubscribe(subscribe.Topic, co.Conn)
	/* d e b u g */
}

// onPING is the heartbeat handler ( other packets reset its timer as well ).
func (co *COnline) OnPING(packet protobase.PacketInterface) {
	logger.Debug("+ [Heartbeat] Received.")
}

// onSUBACK is a handler which removes the outbound subscribe
// message when QoS >0.
func (co *COnline) OnSUBACK(packet protobase.PacketInterface) {
	// TODO
	const fn string = "OnSUBACK"
	var (
		pa       *Suback = NewSuback(packet)
		uid      uuid.UUID
		oidstore protobase.MSGIDInterface
		msgid    uint16
		npc      *Subscribe
		pb       *protocol.MsgBox
		pbc      protobase.MsgInterface
	)
	logger.FInfo(fn, "* [COnline] packet is received.")
	if pa == nil {
		logger.FDebug(fn, "- [Decode] uanble to decode in [SubAck].", packet)
		co.Shutdown()
		return
	}
	// if err := pa.DecodeFrom(packet.GetData()); err != nil {
	// 	logger.FDebug(fn, "- [Decode] uanble to decode in [SubAck].", err)
	// 	return
	// }
	oidstore = co.Conn.storage.GetIDStoreO()
	msgid = pa.Meta.MessageId
	uid, ok := oidstore.GetUUID(msgid)
	if !ok {
		logger.FWarn(fn, "- [COnline][IDStore/Suback] no packet with msgid found.", "msgid", msgid)
		co.Shutdown()
		return
	}
	np, ok := co.Conn.storage.GetOutbound(uid)
	if !ok {
		logger.FWarn(fn, "- [COnline][MessageBox/Suback] no packet with uid found.", uid)
		co.Shutdown()
		return
	}
	if !co.Conn.storage.DeleteOut(np) {
		logger.FWarn(fn, "- [COnline][MessageBox/Suback] failed to remove message.")
		co.Shutdown()
		return
	}
	oidstore.FreeId(msgid)
	npc = np.(*Subscribe)
	if npc == nil {
		// TODO
		// . handle this case
		logger.FWarn(fn, "- *****[COnline][MessageBox/Suback] npc==nil [FATAL].")
	}
	pb = protocol.NewMsgBox(npc.Meta.Qos, npc.Meta.MessageId, protobase.MDInbound, protocol.NewMsgEnvelope(npc.Topic, nil))
	pbc = pb.Clone(protobase.MDInbound)
	/* critical section */
	co.Conn.clblock.Lock()
	callback, ok := co.Conn.clbsub[msgid]
	if ok {
		delete(co.Conn.clbsub, msgid)
	}
	co.Conn.clblock.Unlock()
	/* critical section - end */
	if ok && callback != nil {
		callback(nil, pbc)
		co.client.Subscribe(pbc)
		return
	}
}

// onPUBACK is a handler which removes the outbound publish
// message when QoS >0.
func (co *COnline) OnPUBACK(packet protobase.PacketInterface) {
	// TODO
	var (
		pa       *Puback = protocol.NewPuback(packet)
		uid      uuid.UUID
		oidstore protobase.MSGIDInterface
		msgid    uint16
		pb       *protocol.MsgBox
	)
	logger.FDebug("onPUBACK", "+ [PubAck] packet received.")
	if pa == nil {
		logger.FDebug("onPUBACK", "- [Decode] uanble to decode in [PubAck].", packet)
		co.Shutdown()
		return
	}
	// if err := pa.DecodeFrom(packet.GetData()); err != nil {
	// 	logger.FDebug("onPUBACK", "- [Decode] uanble to decode in [PubAck].", err)
	// 	return
	// }
	oidstore = co.Conn.storage.GetIDStoreO()
	msgid = pa.Meta.MessageId
	uid, ok := oidstore.GetUUID(msgid)
	if !ok {
		logger.FWarn("onPUBACK", "- [IDStore/Puback] no packet with msgid found.", "msgid", msgid)
		co.Shutdown()
		return
	}
	np, ok := co.Conn.storage.GetOutbound(uid)
	if !ok {
		logger.FWarn("onPUBACK", "- [MessageBox/Puback] no packet with uid found.", uid)
		co.Shutdown()
		return
	}
	if !co.Conn.storage.DeleteOut(np) {
		logger.FWarn("onPUBACK", "- [MessageBox/Puback] failed to remove message.")
		co.Shutdown()
		return
	}
	oidstore.FreeId(msgid)
	npc := np.(*Publish)
	if npc == nil {
		// TODO
		// . handle this case
		logger.FWarn("onPUBACK", "- [MessageBox/Puback] npc==nil [FATAL].")
		co.Shutdown()
		return
	}
	pb = protocol.NewMsgBox(npc.Meta.Qos, npc.Meta.MessageId, protobase.MDInbound,
		protocol.NewMsgEnvelope(npc.Topic, npc.Message))
	/* critical section */
	co.Conn.clblock.Lock()
	callback, ok := co.Conn.clbpub[msgid]
	if ok {
		delete(co.Conn.clbpub, msgid)
	}
	co.Conn.clblock.Unlock()
	/* critical section - end */
	if ok && callback != nil {
		callback(nil, pb)
		return
	}
}

func (co *COnline) OnDISCONNECT(packet protobase.PacketInterface) {
	// TODO
	const fn string = "OnDISCONNECT"
	var (
		err error
	)
	logger.FDebug(fn, "+ [COnline] Disconnect packet is received.")
	err = co.Conn.protocon.Conn.Close()
	if err != nil {
		logger.FWarn(fn, "- [COnline] unable to close the connection. error:", err)
	}
}

func (co *COnline) OnPONG(packet protobase.PacketInterface) {
	// TODO
	const fn string = "OnPONG"
	logger.FDebug(fn, "* [COnline] packet received.")
}

func (co *COnline) OnQueue(packet protobase.PacketInterface) {
	const fn string = "OnQueue"
	logger.FDebug(fn, "+ [COnline] packet is received.")
}

func (co *COnline) OnQueueAck(packet protobase.PacketInterface) {
	const fn string = "OnQueueAck"
	logger.FDebug(fn, "+ [COnline] packet is received.")
}

// Shutdown sets the status to error which notifies the supervisor
// and cleanly terminates the connection.
func (co *COnline) Shutdown() {
	const fn string = "Shutdown"
	logger.FDebug(fn, "* [COnline] closing.")
	atomic.StoreUint32(&(co.Conn).Status, STATERR)
}
