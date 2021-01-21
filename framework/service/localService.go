package service

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"sync"
	"time"

	"github.com/giant-tech/go-service/base/serializer"
	"github.com/giant-tech/go-service/framework/idata"
	"github.com/giant-tech/go-service/framework/iserver"
	"github.com/giant-tech/go-service/framework/msgdef"
	"github.com/giant-tech/go-service/framework/servicedef"

	"github.com/cihub/seelog"
	"github.com/spf13/viper"
)

// LocalService 本进程服务
type LocalService struct {
	iserver.IBaseCtrlService
	// dataC 是其他服务向本服务发送消息的Channel，协程中将依次执行动作。
	dataC chan *idata.CallData

	closeSig chan bool
	wg       sync.WaitGroup
}

// createLocalService 创建本地服务
func createLocalService(sname string) (*LocalService, error) {
	ls := &LocalService{}

	err := ls.InitLocalService(sname)
	if err != nil {
		return nil, err
	}

	return ls, nil
}

// InitLocalService 初始化本地服务
func (s *LocalService) InitLocalService(sname string) error {
	is := GetServiceByName(sname)
	if is == nil {
		seelog.Error("can't find service: ", sname)
		return fmt.Errorf("service not found: %s", sname)
	}

	s.dataC = make(chan *idata.CallData, 10240)
	s.closeSig = make(chan bool, 1)
	s.IBaseCtrlService = reflect.New(reflect.TypeOf(is.ServiceCtrl).Elem()).Interface().(iserver.IBaseCtrlService)

	err := s.IBaseCtrlService.InitBaseService(sname, is.ServiceTypeID, s)
	if err != nil {
		seelog.Error("InitBaseService ", sname, " error: ", err)
		return err
	}

	//注册RPC
	s.IBaseCtrlService.RegRPCMsg(s.IBaseCtrlService)
	err = s.IBaseCtrlService.OnInit()
	if err == nil {
		//检测lobby service是否已实现所有lobbyservice.json里描述的接口
		rpchandlers, err := s.IBaseCtrlService.GetRPCHandlers()
		if err == nil {
			seelog.Debug("service check : ", sname)
			def := servicedef.GetServiceDefs().GetDef(sname)
			if def != nil {
				for methodname := range def.Methods {
					_, ok := rpchandlers.Load(methodname)
					if !ok {
						seelog.Error("name: ", sname, " method: ", methodname, " not implement")
					}
				}
			}
		}

		//参数检测
		methodsparams, err := s.IBaseCtrlService.GetRPCMethodParams()
		def := servicedef.GetServiceDefs().GetDef(sname)
		if err == nil {
			if def != nil {
				for methodname, jsonParams := range def.MethodsParams {
					params, ok := methodsparams.Load(methodname)
					if ok && jsonParams != nil && params != nil {
						if !reflect.DeepEqual(jsonParams, params) {
							seelog.Error("name: ", sname, " method: ", methodname, ", params not equal", " jsonParams = ", jsonParams, " params = ", params)
						}
					}
				}
			}
			seelog.Debug("service check success: ", sname)
		}
	}
	return err
}

// PostFunction 投递函数给实体，并在实体所在的协程中执行
func (s *LocalService) PostFunction(f func()) {
	s.dataC <- &idata.CallData{Func: f}
}

// PostFunctionAndWait 投递函数给实体协程执行，并等待执行结果
func (s *LocalService) PostFunctionAndWait(f func() interface{}) interface{} {
	// 结果从ch返回
	ch := make(chan interface{}, 1)
	s.dataC <- &idata.CallData{Func: func() { ch <- f() }}

	// 等待直到返回结果
	return <-ch
}

// PostCallMsg 投递消息给本服务，立即返回
func (s *LocalService) PostCallMsg(msg *msgdef.CallMsg) error {
	// seelog.Infof("PostCallMsg, Seq:%d, MethodName:%s, groupID:%d, entityID:%d",
	// 	msg.Seq, msg.MethodName, msg.GroupID, msg.EntityID)

	// 发送给客户端，直接转发
	if msg.SType == uint8(idata.ServiceClient) && msg.EntityID != 0 {
		e := s.GetEntity(msg.EntityID)
		if e == nil || e.GetClientSess() == nil {
			seelog.Error("LocalService.PostCallMsg, entity not found")
			return fmt.Errorf("entity not found")
		}

		return e.GetClientSess().Send(msg)
	}

	//如果是多线程，实体的消息直接给实体管道
	if s.IsMultiThread() && msg.EntityID != 0 {
		var e iserver.IEntity
		if msg.GroupID != 0 {
			g := s.GetEntity(msg.GroupID)
			if g == nil {
				seelog.Error("group entity not found, groupID: ", msg.GroupID)
				return fmt.Errorf("group entity not found, groupID: %d", msg.GroupID)
			}

			es, ok := g.(iserver.IEntities)
			if !ok {
				seelog.Error("entity is not IEntities, groupID: ", msg.GroupID)
				return fmt.Errorf("entity is not IEntities, groupID: %d", msg.GroupID)
			}

			e = es.GetEntity(msg.EntityID)
		} else {
			e = s.GetEntity(msg.EntityID)
		}

		if e == nil {
			seelog.Error("entity not found: ", msg.EntityID)
			return fmt.Errorf("entity not found")
		}

		return e.PostCallMsg(msg)
	}
	data := &idata.CallData{}
	data.Msg = msg

	s.dataC <- data

	return nil

}

// PostCallMsgAndWait 投递消息给本服务，并等待结果返回
func (s *LocalService) PostCallMsgAndWait(msg *msgdef.CallMsg) *idata.RetData {
	// 发送给客户端的协议需要特殊处理
	if msg.SType == uint8(idata.ServiceClient) && msg.EntityID != 0 {
		data := &idata.RetData{}
		data.Err = fmt.Errorf("not suport entity SyncCall client method")
		seelog.Error("LocalService.PostCallMsgAndWait, not suport entity SyncCall client method")
		return data
	}

	//如果是多线程，实体的消息直接给实体管道
	if s.IsMultiThread() && msg.EntityID != 0 {
		var e iserver.IEntity
		if msg.GroupID != 0 {
			g := s.GetEntity(msg.GroupID)
			if g != nil {
				es, ok := g.(iserver.IEntities)
				if ok {
					e = es.GetEntity(msg.EntityID)
				}
			}
		} else {
			e = s.GetEntity(msg.EntityID)
		}

		if e == nil {
			data := &idata.RetData{}
			data.Err = fmt.Errorf("entity not found")

			return data
		}

		return e.PostCallMsgAndWait(msg)
	}
	data := &idata.CallData{}
	data.Msg = msg

	// 结果从ChanRet返回
	data.ChanRet = make(chan *idata.RetData, 1)
	s.dataC <- data

	// 等待直到返回结果
	return <-data.ChanRet

}

// Run 服务开始
func (s *LocalService) Run(closeSig chan bool) {
	seelog.Debug("run service , serviceName: ", s.GetSName(), ", serviceType: ", s.GetSType(), ", ServerID: ", s.GetSID(), " tickMS: ", s.GetTickMS())

	//通知上层服务可用
	localservices := GetLocalServiceMgr().GetAllLocalServiceForConnectNotify(s.GetSID(), s.GetSType())

	//如果配了notconnect service, 就无需回调OnConnected
	if len(localservices) > 0 {
		data := serializer.SerializeNew(localservices)
		msg := &msgdef.CallMsg{
			SType:      uint8(s.GetSType()),
			MethodName: "Connected",
			Params:     data,
		}

		s.PostCallMsg(msg)
	}

	ticker := time.NewTicker(time.Duration(s.GetTickMS()) * time.Millisecond)
	defer ticker.Stop()

	loopFun := func() bool {
		defer func() {
			if err := recover(); err != nil {
				seelog.Error("LocalService Run panic:", err, ", ", string(debug.Stack()))
				if viper.GetString("Config.Recover") == "0" {
					panic(err)
				}
			}
		}()

		select {
		case <-closeSig:
			return false

		case <-ticker.C:
			s.OnTick()

		case callData := <-s.dataC:
			s.processCall(callData)
		}

		return true
	}

	for loopFun() {
	}
}

func (s *LocalService) processCall(data *idata.CallData) {
	//如果Func不为nil则直接调用
	if data.Func != nil {
		data.Func()
		return
	}

	if data.Msg.EntityID != 0 {
		var e iserver.IEntity
		if data.Msg.GroupID != 0 {
			g := s.GetEntity(data.Msg.GroupID)
			if g != nil {
				es, ok := g.(iserver.IEntities)
				if ok {
					e = es.GetEntity(data.Msg.EntityID)
				}
			}
		} else {
			e = s.GetEntity(data.Msg.EntityID)
		}

		if e != nil {
			e.DoRPCMsg(data.Msg.MethodName, data.Msg.Params, data.ChanRet)
		} else {
			seelog.Error("processCall, entity not exist,methodName: ", data.Msg.MethodName, ", groupID: ", data.Msg.GroupID, ", entityID: ", data.Msg.EntityID)
			//实体不存在
			if data.ChanRet != nil {
				//同步调用
				retData := &idata.RetData{}
				retData.Err = fmt.Errorf("entity not exist: %d", data.Msg.EntityID)
				data.ChanRet <- retData
			} else {
				//异步调用失败暂时没有消息通知
			}
		}
	} else {
		s.DoRPCMsg(data.Msg.MethodName, data.Msg.Params, data.ChanRet)
	}
}
